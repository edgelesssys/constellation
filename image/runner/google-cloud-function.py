import os
import sys
import re
import hmac
import hashlib
import random
import string
import google.cloud.compute_v1 as compute_v1

LABEL="nested-virt"
AUTH_TOKEN_ENV="COREOS_BUILDER_WORKFLOW_FUNCTION_TOKEN"
SA_EMAIL="constellation-cos-builder@constellation-331613.iam.gserviceaccount.com"
SA_SCOPES=[
    "https://www.googleapis.com/auth/compute",
    "https://www.googleapis.com/auth/servicecontrol",
    "https://www.googleapis.com/auth/cloud-platform",
]

def workflow_job(request):
    """Responds to https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#workflow_job
    Args:
        request (flask.Request): HTTP request object.
    Returns:
        The response text or any set of values that can be turned into a
        Response object using
        `make_response <http://flask.pocoo.org/docs/1.0/api/#flask.Flask.make_response>`.
    """
    allow, reason = authorize(request)
    if not allow:
        return f'unauthorized: {reason}'
    request_json = request.get_json()
    if request_json and 'action' in request_json:
        if request_json['action'] == 'queued':
            return job_queued(request_json['workflow_job'])
        elif request_json['action'] == 'completed':
            return job_completed(request_json['workflow_job'])
        elif request_json['action'] == 'in_progress':
            return f'nothing to do here'
    else:
        return f'invalid message format'

def authorize(request) -> (bool, str) :
    correct_token = os.environ.get(AUTH_TOKEN_ENV)
    if correct_token is None:
        return False, 'correct token not set'
    correct_hmac = 'sha256=' + hmac.new(correct_token.encode('utf-8'), request.get_data(), hashlib.sha256).hexdigest()
    request_hmac = request.headers.get('X-Hub-Signature-256')
    if request_hmac is None:
        return False, 'X-Hub-Signature-256 not set'
    if correct_hmac == request_hmac:
        return True, ''
    else:
        return False, f'X-Hub-Signature-256 incorrect'


def job_queued(workflow_job) -> str:
    if not LABEL in workflow_job['labels']:
        return f'unexpected job labels: {workflow_job["labels"]}'
    cloud_init = generate_cloud_init()
    instance_uid = ''.join(random.choice(string.ascii_lowercase + string.digits) for i in range(6))
    try:
        create_instance(metadata={'user-data': cloud_init}, instance_name=f'coreos-builder-{instance_uid}')
    except Exception as e:
        return f'creating instance failed: {e}'
    return 'success'

def job_completed(workflow_job) -> str:
    if not LABEL in workflow_job['labels']:
        return f'unexpected job labels: {workflow_job["labels"]}'
    instance_name = workflow_job["runner_name"]
    try:
        delete_instance(machine_name=instance_name)
    except Exception as e:
        return f'deleting instance failed: {e}'
    return 'success'

def generate_cloud_init() -> str:
    with open("cloud-init.txt", "r") as f:
        cloud_init = f.read()
    return cloud_init

def create_instance(
    metadata: dict[str, str],
    project_id: str = 'constellation-331613',
    zone: str = 'us-central1-c',
    instance_name: str = 'coreos-builder',
    machine_type: str = "n2-highmem-4",
    source_image: str = "projects/ubuntu-os-cloud/global/images/family/ubuntu-2004-lts",
    network_name: str = "global/networks/default",
    disk_size_gb: int = 64,
    enable_nested_virtualization: bool = True,
    service_accounts: list[compute_v1.ServiceAccount] = [compute_v1.ServiceAccount(email=SA_EMAIL, scopes=SA_SCOPES)],
) -> compute_v1.Instance:
    """
    Send an instance creation request to the Compute Engine API and wait for it to complete.

    Args:
        project_id: project ID or project number of the Cloud project you want to use.
        zone: name of the zone you want to use. For example: “us-west3-b”
        instance_name: name of the new virtual machine.
        machine_type: machine type of the VM being created. This value uses the
            following format: "zones/{zone}/machineTypes/{type_name}".
            For example: "zones/europe-west3-c/machineTypes/f1-micro"
        source_image: path to the operating system image to mount on your boot
            disk. This can be one of the public images
            (like "projects/debian-cloud/global/images/family/debian-10")
            or a private image you have access to.
        network_name: name of the network you want the new instance to use.
            For example: "global/networks/default" represents the `default`
            network interface, which is created automatically for each project.
    Returns:
        Instance object.
    """
    instance_client = compute_v1.InstancesClient()
    operation_client = compute_v1.ZoneOperationsClient()

    # Describe the size and source image of the boot disk to attach to the instance.
    disk = compute_v1.AttachedDisk()
    initialize_params = compute_v1.AttachedDiskInitializeParams()
    initialize_params.source_image = (
        source_image
    )
    initialize_params.disk_size_gb = disk_size_gb
    disk.initialize_params = initialize_params
    disk.auto_delete = True
    disk.boot = True
    disk.type_ = "PERSISTENT"

    # Use the network interface provided in the network_name argument.
    network_interface = compute_v1.NetworkInterface()
    network_interface.name = network_name
    network_interface.access_configs = [compute_v1.AccessConfig()]

    # Collect information into the Instance object.
    instance = compute_v1.Instance()
    instance.name = instance_name
    instance.disks = [disk]
    if re.match(r"^zones/[a-z\d\-]+/machineTypes/[a-z\d\-]+$", machine_type):
        instance.machine_type = machine_type
    else:
        instance.machine_type = f"zones/{zone}/machineTypes/{machine_type}"
    instance.network_interfaces = [network_interface]

    # Enable nested virtualization if requested
    advanced_machine_features = compute_v1.AdvancedMachineFeatures()
    advanced_machine_features.enable_nested_virtualization = enable_nested_virtualization
    instance.advanced_machine_features = advanced_machine_features

    metadata_items = [compute_v1.Items(key=k, value=v) for k, v in metadata.items()]
    metadata = compute_v1.Metadata(items=metadata_items)
    instance.metadata = metadata

    # set service accounts.
    instance.service_accounts = service_accounts

    # Prepare the request to insert an instance.
    request = compute_v1.InsertInstanceRequest()
    request.zone = zone
    request.project = project_id
    request.instance_resource = instance

    # Wait for the create operation to complete.
    print(f"Creating the {instance_name} instance in {zone}...")
    operation = instance_client.insert_unary(request=request)
    while operation.status != compute_v1.Operation.Status.DONE:
        operation = operation_client.wait(
            operation=operation.name, zone=zone, project=project_id
        )
    if operation.error:
        print("Error during creation:", operation.error, file=sys.stderr)
    if operation.warnings:
        print("Warning during creation:", operation.warnings, file=sys.stderr)
    print(f"Instance {instance_name} created.")
    return instance

def delete_instance(
    project_id: str = 'constellation-331613',
    zone: str = 'us-central1-c',
    machine_name: str = 'coreos-builder',
    ) -> None:
    """
    Send an instance deletion request to the Compute Engine API and wait for it to complete.

    Args:
        project_id: project ID or project number of the Cloud project you want to use.
        zone: name of the zone you want to use. For example: “us-west3-b”
        machine_name: name of the machine you want to delete.
    """
    instance_client = compute_v1.InstancesClient()
    operation_client = compute_v1.ZoneOperationsClient()

    print(f"Deleting {machine_name} from {zone}...")
    operation = instance_client.delete_unary(
        project=project_id, zone=zone, instance=machine_name
    )
    while operation.status != compute_v1.Operation.Status.DONE:
        operation = operation_client.wait(
            operation=operation.name, zone=zone, project=project_id
        )
    if operation.error:
        print("Error during deletion:", operation.error, file=sys.stderr)
    if operation.warnings:
        print("Warning during deletion:", operation.warnings, file=sys.stderr)
    print(f"Instance {machine_name} deleted.")
    return
