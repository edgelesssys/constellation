# Publishing Marketplace Images

Constellation release images need to be manually published to AWS and Azure marketplaces due to the lack of automation features.
On GCP, marketplace image publishing is automated and takes place on release.

This document explains how to perform the uploading on AWS and Azure.

## AWS

1. Log in to the [AWS marketplace management portal](https://aws.amazon.com/marketplace/management/) with your regular developer AWS account.
2. Select "Products -> Server -> Constellation" in the top menu.
3. Select "Versions" in the main menu and press "Add version".
4. Fill in the form.
    1. Enter the semantic version of the release (i.e. `vX.Y.Z`) as "Version title".
    2. Set the version tag in "Release notes" to the same version.
    3. For the "Amazon Machine Image (AMI) ID", enter the AMI ID of the release (SEV-SNP) image. This can be found in the regular
       [AWS console](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#Images:visibility=owned-by-me;search=:constellation-v;v=3;$case=tags:false%5C,client:false;$regex=tags:false%5C,client:false;sort=desc:creationDate).
    4. For "IAM access role ARN", enter `arn:aws:iam::795746500882:role/constellation-marketplace-ingest`.
5. Leave the other fields as they are and press "Add version".
6. Wait for the [request](https://aws.amazon.com/marketplace/management/requests) to be processed and available before publishing the release.

## Azure

1. Log in to the [Microsoft partner center](https://partner.microsoft.com/en-us/dashboard/home) with your regular developer Microsoft account.
2. Select "Marketplace offers -> Constellation -> Constellation Node" in the main menu.
3. Select "Technical configuration" in the sidebar on the left.
4. Select "Add VM Image".
    1. For the "Version number", enter the semantic version of the release without the `v` prefix. If the release version is `vX.Y.Z`, enter `X.Y.Z`.
    2. Press "Add a gallery image" and select the corresponding "Constellation_CVM" image version in the menu.
    3. Press "Save VM image".
5. **IMPORTANT**: Hit **Save draft**. Do **NOT** hit "Review and publish" directly.
6. **After** saving the draft, hit "Review and publish".
7. Go back to the [offer's home page](https://partner.microsoft.com/en-us/dashboard/commercial-marketplace/offers/a53ac90b-06f7-4a20-a845-8607ca352e61/overview) and wait for the process to complete before publishing the release.
