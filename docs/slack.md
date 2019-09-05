## Slack configurations guidelines

### **Step 1**: Create a public HTTPS URL for Comedian
Install [ngrok](https://ngrok.com/product) and create a public HTTPS URL for Comedian on your development machine by running `ngrok http 8080`.

### **Step 2**: Create Slack chatbot 
Create "app" in slack workspace: https://api.slack.com/apps. Obtain App Credentals in `basic info` section and export them

```
export SLACK_CLIENT_ID=383672116036.563661723157
export SLACK_CLIENT_SECRET=6b0826c3b77fd072dc1ec1fc5c582743
export SLACK_VERIFICATION_TOKEN=Oiwpp2x5Jup1jdQxdtnYTOWT
```

### **Step 3**: Add bot user 
From the left sidebar select "Bot users". Create a bot user with any name you like. Turn on "Always show my bot online" feature. 

### **Step 4**: Configure slash commands
From the left sidebar select "Slash Commands". Create slash command with request URL: `http://<your ngrok https URL>/commands`) Mark as needed the option of `Escase channels, users and links sent to your app`. 

| Name | Hint | Description |
| --- | --- | --- | 
| /start | start standuping with role | Adds a new user with selected role  |
| /quit | - | Removes user with selected role from standup team |
| /show | - | Shows users assigned to standup in the current chat |
| /show_deadline | - | Show standup time in current channel |
| /update_deadline | - | Update or delete standup time in current channel |

### **Step 5**: Add Redirect URL in OAuth & Permissions tab
Add a new redirect url `http://<ngrok https URL>/auth`. Save it! This is where Slack will redirect when you install bot into a workspace

### **Step 7**: Add Event Subscriptions
Run Comedian with `make run` command 

In Event Subscriptions tab enable events. Configure URL as follows ```http://<ngrok https URL>/event```. You should receive confirmation of your endpoint. if not, check if Comedian and ngrok are up and working and you have internet access. If confirm received, add `app_uninstalled`, `message_groups`, `message_channels`, `team_join` events. 

### **Step 8**: Add Comedian to your workspace
Navigate to `manage distribution` tab and press `Add to Slack` button
Chose which slack you want to add Comedian to and authorize it.