<div align="center">
    <img style="width: 300px" src ="logo.png" />
</div>

<div align="center"> Team management system that helps track performance and assist team members in daily remote standups meetings 

[![Developed by Mad Devs](https://maddevs.io/badge-dark.svg)](https://maddevs.io/)
[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
[![Go Report Card](https://goreportcard.com/badge/github.com/maddevsio/comedian)](https://goreportcard.com/report/github.com/maddevsio/comedian)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

## Comedian Features

- [x] Handle standups and show warnings if standup is not complete 
- [x] Assign team members the roles of project managers, developers, designers and testers
- [x] Set deadlines for standups submissions in channels
- [x] Set up individual timetables (schedules) for developers to submit standups
- [x] Remind about upcoming deadlines for teams and individuals
- [x] Tag non-reporters in channels when deadline is missed
- [x] Provide daily & weekly reports on team's performance
- [x] Support English and Russian languages

## Local Development
Comedian works with Slack apps, if you already have a slack app configured for comedian, proceed with step, othervise, follow the steps below: 

### **Step 1**: Create a public HTTPS URL for Comedian
Install [ngrok](https://ngrok.com/product) and create a public HTTPS URL for Comedian on your development machine by running `ngrok http 8080`

### **Step 2**: Create Slack chatbot 
Create "app" in slack workspace: https://api.slack.com/apps
In the drop-down list at the top select the created "app"
Obtain App Credentals in basic info section and export them

```
export SLACK_CLIENT_ID=383672116036.563661723157
export SLACK_CLIENT_SECRET=6b0826c3b77fd072dc1ec1fc5c582743
export SLACK_VERIFICATION_TOKEN=Oiwpp2x5Jup1jdQxdtnYTOWT
```

### **Step 3**: Add bot user 
From the left sidebar select "Bot users". Create a bot user with any name you like. Turn on "Always show my bot online" feature. 

### **Step 4**: Configure slash commands
From the left sidebar select "Slash Commands". Create `/comedian` slash command with request URL: `http://<ngrok https URL>/commands`) Mark as needed the option of `Escase channels, users and links sent to your app`. 

### **Step 5**: Add Redirect URL in OAuth & Permissions tab
Add a new redirect url `http://<ngrok https URL>/auth`. Save it! This is where Slack will redirect when you install bot into a workspace

### **Step 6**: Run Comedian

CD into your project root directory and run Comedian with `make run` command from your terminal. In case you do not have docker and docker-compose, install them on your machine and try again. 

### **Step 7**: Add Event Subscriptions
In Event Subscriptions tab enable events. Configure URL as follows ```http://<ngrok https URL>/event```. You should receive confirmation of your endpoint. if not, check if your app and ngrok are working and you have internet access. If confirm received, add `app_uninstalled`, `message_groups`, `message_channels`, `team_join` events. 

### **Step 8**: Add Comedian to your workspace
Navigate to `manage distribution` tab and press `Add to Slack` button
Chose which slack you want to add Comedian to and authorize it. Once done, Comedian will send you the username and password. 

## Usage

You can use the following commands:
***
| Name | Hint | Description |
| --- | --- | --- | 
| /comedian | | displays helpful info about all commands |
| /comedian add | @user @user1 / (admin, pm, developer) | Adds a new user with selected role |
| /comedian remove | @user @user1 / (admin, pm, developer) | Removes user with selected role |
| /comedian show | - | Shows users assigned to standup |
| /comedian add_deadline | hh:mm | Set standup time |
| /comedian show_deadline | - | Show standup time in current channel |
| /comedian remove_deadline | - | Delete standup time in current channel |

To update messages: 
```
goi18n extract
goi18n merge active.*.toml
goi18n merge active.*.toml translate.*.toml
```

## Testing

Run tests with `make test` command. This will run integration tests and output the result.

If you want to do manual testing for separate components / or see code coverage with `vscode` or `go test`, use `make setup` to setup database for testing purposes.