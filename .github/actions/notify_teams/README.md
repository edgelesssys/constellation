# notify Teams action

This action is used to send a message to our Teams channel in case of a failure in the CI/CD pipeline.
The action will automatically choose an engineer to assign to the issue and tag them in the message.

Engineers are identified by their GitHub username and bound to a Microsoft Teams ID in `.attachments[0].content.msteams.entities`.
To add a new engineer, add a new entry to the entity list in the format:

```json
{
  "type": "mention",
  "text": "${github_username}",
  "mentioned": {
    "id": "${msteams_id}",
    "name": "${name}"
  }
}
```

Where `${github_username}` is the GitHub username of the engineer, `${msteams_id}` is the Microsoft Teams ID of the engineer, and `${name}` is the name of the engineer.
To find the Microsoft Teams ID use the following command:

```bash
az ad user show --id ${email} --query id
```

Where `${email}` is the email address of the engineer.
