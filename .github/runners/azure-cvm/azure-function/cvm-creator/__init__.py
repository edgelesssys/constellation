import os
import logging
import hmac
import hashlib
import random
import string
import base64
from typing import Tuple

import azure.functions as func

from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.resource.resources.v2021_04_01.models import Deployment, DeploymentProperties
from azure.keyvault.secrets import SecretClient
from azure.identity import DefaultAzureCredential

LABEL = "azure-cvm"
SUBSCRIPTION_ID = "0d202bbb-4fa7-4af8-8125-58c269a05435"
RESOURCE_GROUP = "snp-value-reporting"
VAULT_URL = "https://github-token.vault.azure.net/"
TOKEN_SECRET_NAME = "gh-webhook-secret"
SSH_KEY_SECRET_NAME = "snp-reporter-pubkey"


def main(req: func.HttpRequest) -> func.HttpResponse:
    logging.info('Python HTTP trigger function processed a request.')

    allow, reason = authorize(req)
    if not allow:
        return func.HttpResponse(f'unauthorized: {reason}', status_code=401)

    request_json = req.get_json()
    if request_json and 'action' in request_json:
        if request_json['action'] == 'queued':
            return job_queued(request_json['workflow_job'])
        elif request_json['action'] == 'completed':
            return job_completed(request_json['workflow_job'])
        elif request_json['action'] == 'in_progress':
            return f'nothing to do here'
    else:
        return func.HttpResponse(f'invalid message format', status_code=400)


def authorize(request) -> Tuple[bool, str]:
    credentials = DefaultAzureCredential()
    client = SecretClient(vault_url=VAULT_URL, credential=credentials)
    correct_token = client.get_secret(TOKEN_SECRET_NAME).value

    if correct_token is None:
        return False, 'correct token not set'
    correct_hmac = 'sha256=' + \
        hmac.new(correct_token.encode('utf-8'),
                 request.get_body(), hashlib.sha256).hexdigest()
    request_hmac = request.headers.get('X-Hub-Signature-256')

    if request_hmac is None:
        return False, 'X-Hub-Signature-256 not set'
    if hmac.compare_digest(correct_hmac, request_hmac):
        return True, ''
    else:
        return False, f'X-Hub-Signature-256 incorrect'


def job_queued(workflow_job) -> str:
    if not LABEL in workflow_job['labels']:
        return func.HttpResponse(f'irrelevant job labels: {workflow_job["labels"]}', status_code=200)
    cloud_init = generate_cloud_init()
    instance_uid = ''.join(random.choice(
        string.ascii_lowercase + string.digits) for i in range(6))

    credentials = DefaultAzureCredential()
    client = SecretClient(vault_url=VAULT_URL, credential=credentials)
    ssh_key = client.get_secret(SSH_KEY_SECRET_NAME).value

    try:
        create_cvm(instance_uid, cloud_init, ssh_key)
    except Exception as e:
        return func.HttpResponse(f'creating instance failed: {e}', status_code=400)
    return 'success'


def job_completed(workflow_job) -> str:
    if not LABEL in workflow_job['labels']:
        return func.HttpResponse(f'irrelevant job labels: {workflow_job["labels"]}', status_code=200)
    instance_name = workflow_job["runner_name"]
    try:
        delete_cvm(machine_name=instance_name)
    except Exception as e:
        return func.HttpResponse(f'deleting instance failed: {e}', status_code=400)
    return 'success'


def generate_cloud_init() -> str:
    path = os.path.join(os.path.dirname(__file__), "cloud-init.txt")
    with open(path, "r") as f:
        cloud_init = f.read()

    return base64.b64encode(cloud_init.encode('utf-8'))


def delete_cvm(machine_name):
    credentials = DefaultAzureCredential()
    resource_client = ResourceManagementClient(
        credentials,
        SUBSCRIPTION_ID,
    )

    path = f"/subscriptions/{SUBSCRIPTION_ID}/resourceGroups/{RESOURCE_GROUP}/providers"

    async_vm_delete = resource_client.resources.begin_delete_by_id(
        resource_id=f"{path}/Microsoft.Compute/virtualMachines/{machine_name}", api_version="2022-08-01")
    async_vm_delete.wait()
    async_osdisk_delete = resource_client.resources.begin_delete_by_id(
        resource_id=f"{path}/Microsoft.Compute/disks/{machine_name}-osdisk", api_version="2022-07-02")
    async_nic_delete = resource_client.resources.begin_delete_by_id(
        resource_id=f"{path}/Microsoft.Network/networkInterfaces/{machine_name}-nic", api_version="2022-08-01")
    async_nsg_delete = resource_client.resources.begin_delete_by_id(
        resource_id=f"{path}/Microsoft.Network/networkSecurityGroups/{machine_name}-nsg", api_version="2022-05-01")
    async_vnet_delete = resource_client.resources.begin_delete_by_id(
        resource_id=f"{path}/Microsoft.Network/virtualNetworks/{machine_name}-vnet", api_version="2022-05-01")
    async_ip_delete = resource_client.resources.begin_delete_by_id(
        resource_id=f"{path}/Microsoft.Network/publicIPAddresses/{machine_name}-ip", api_version="2022-05-01")

    async_vnet_delete.wait()
    async_nic_delete.wait()
    async_ip_delete.wait()
    async_nsg_delete.wait()
    async_osdisk_delete.wait()

    return True


def create_cvm(instance_uid, cloud_init, ssh_key) -> str:
    credentials = DefaultAzureCredential()
    resource_client = ResourceManagementClient(
        credentials,
        SUBSCRIPTION_ID,
    )

    template_id = "https://raw.githubusercontent.com/edgelesssys/constellation/main/.github/runners/azure-cvm/cvm-template.json"

    depl_properties = DeploymentProperties(mode="Incremental", template_link={"uri": template_id}, parameters={"instanceUid": {
                                           "value": instance_uid}, "customData": {"value": cloud_init.decode("utf-8")}, "pubKey": {"value": ssh_key}})
    depl = Deployment(properties=depl_properties)

    async_vm_start = resource_client.deployments.begin_create_or_update(
        "snp-value-reporting", "snp-value-reporter-deployment", depl)

    async_vm_start.wait()

    return True
