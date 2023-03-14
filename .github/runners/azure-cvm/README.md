# General
This folder contains the files to setup an Azure function and ARM template in order to deploy Azure CVMs with a webhook.

- `cvm-template.json`: An ARM template that deploys one CVM and the required resources. It is deployed by the Azure Function
- `azure-function`: All necessary files to redeploy the function. Changes in `requirements.txt` are installed during deployment of the function. `cloud-init.txt` is put into the CVM by supplying it as a parameter to the ARM template deployment.

# Update cvm-template
In order to make the Azure function use your changes you will have to publish them in the `cvm-template.json` file.
While developing you can point the `template_id` variable in `__init__.py` to a different location.

Doing the following you can debug your template changes:
- Look for the `Template spec` resource in your Azure project (e.g. "snp-value-reporter-template").
- Click on "Create new version".
- Select the latest version available.
- Use `current_version+1` as new version.
- Go to "Edit template" and make your changes.
- Go to "Review + Save" and save your changes.

# Deploy azure function
Background info can be found in the [Azure docs](https://learn.microsoft.com/en-us/azure/azure-functions/create-first-function-cli-python?tabs=azure-cli%2Cbash%2Cbrowser#deploy-the-function-project-to-azure).
To deploy your Azure CLI needs to be authenticated and [Azure Function Core Tools](https://learn.microsoft.com/en-us/azure/azure-functions/functions-run-local?tabs=v4%2Clinux%2Ccsharp%2Cportal%2Cbash#v2) needs to be installed.

```bash
cd .github/runners/azure-cvm/azure-function
func azure functionapp publish edgeless-snp-reporter --python
```
