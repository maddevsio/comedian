<div align="center">
    <img style="width: 300px" src ="documentation/logo.png" />
</div>

<div align="center"> Simple stand up bot for slack to help you with remote stand up meetings automation </div>

<div align="center">

![](https://travis-ci.org/maddevsio/comedian.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/maddevsio/comedian/badge.svg)](https://coveralls.io/github/maddevsio/comedian)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/team-monitoring/comedian)](https://goreportcard.com/report/gitlab.com/team-monitoring/comedian)
[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)
[![](https://godoc.org/gitlab.com/team-monitoring/comedian?status.svg)](https://godoc.org/gitlab.com/team-monitoring/comedian)

</div>

![](documentation/show.gif)

## Comedian Features

- [x] Handle standup and show warnings if standup is not complete 
- [x] Assign developers, Project Managers, and admins
- [x] Control standup deadlines in channels
- [x] Set up individual timetables (schedules) for developers to submit standups
- [x] Remind about upcoming deadlines for teams and individuals
- [x] Tag non-reporters in channels and DM them when deadline is missed
- [x] Generate reports on projects, users or users in projects
- [x] Provide daily report on team's yesterday performance
- [x] Support English and Russian languages


## Getting started

These instructions will help you set up the project on your local machine for development and testing purposes with [ngrok](https://ngrok.com/product) 


### **Step 1**: Create a public HTTPS URL for Comedian
Install [ngrok](https://ngrok.com/product) and create a public HTTPS URL for Comedian on your development machine following the instruction from the web

### **Step 2**: Clone the project
Copy the project repository to your local machine. Note: Go should be already installed! If you do not have Go installed, please, follow [installation guidelines](https://golang.org/doc/install) from Go official website to install it and then proceed to Step 2

```
mkdir -p $GOPATH/src/gitlab.com/team-monitoring/
cd $GOPATH/src/gitlab.com/team-monitoring/
git clone https://gitlab.com/team-monitoring/comedian
```

### **Step 3**: Configure environmental variables

Create `.env` file in the root directory and add the following env variables there. See .env.example for a reference:

| Title | Description | Default | Optional? |
| --- | --- | --- | --- |
| COMEDIAN_SLACK_TOKEN | Bot User OAuth Access Token |  | No |
| COMEDIAN_DATABASE | Database URL. Default: comedian:comedian@/comedian?parseTime=true |  | No |
| COMEDIAN_HTTP_BIND_ADDR | HTTP bind address | 0.0.0.0:8080 | No |
| COMEDIAN_LANGUAGE | Comedian primary language | en_US | No |
| COMEDIAN_SUPER_ADMIN_ID | Slack ID of super admin in your workspace |  | No |
| COMEDIAN_REPORT_CHANNEL | Slack Channel ID to send daily reports to |  | No |
| COMEDIAN_REPORT_TIME | Time to send daily reports | 10:00 | No |
| COMEDIAN_MAX_REMINDERS | Number of times comedian keeps reminding non reporters | 3 | No |
| COMEDIAN_REMINDER_INTERVAL | Duration of the intervals when Comedian waits before next reminder in minutes | 30 | No |
| COMEDIAN_WARNING_TIME | Duration prior to deadline to remind about upcoming deadline | 10 | No |
| TZ | Setup time zone for comedian DB | UTC | Yes |

### **Step 4**: Create Slack chatbot 
Create "app" in slack workspace: https://api.slack.com/apps
In the drop-down list at the top select the created "app"

### **Step 5**: Configure slash commands
In the menu, select "Slash Commands". Create the slash command: ```/comedian```
(Request URL for command: ```http://<ngrok https URL>/commands(here you can paste COMEDIAN_SECRET_TOKEN if it is not empty) ``` )

### **Step 6**: Create bot user
Select "Bot users" in the menu.
Create a new bot user.

### **Step 7**: Install Comedian to your workspace
Go to "OAuth & Permissions"
Press "Install App" button. Authorize bot

### **Step 8**: Configure permissions
In "OAuth & Permissions" tab, scroll down to Scopes section and add additional permission scopes for Comedian:
```
Access information about your workspace
Add slash commands and add actions to messages (and view related content)
Access the workspace's emoji
Send messages as user
Send messages as TestComedian
```

Press "Save Changes" and Reinstall App

### **Step 9**: Finish configuring environmental variables
- Copy Bot User OAuth Access Token (begins with "xoxb") and assign it as COMEDIAN_SLACK_TOKEN env variable
- Copy your Slack User ID and assign it as COMEDIAN_SUPER_ADMIN_ID
- Create a separate channel for reporting and copy its ID, assign it to COMEDIAN_REPORT_CHANNEL

Run the following commands to update your env variables: 
```
set -a
. .env
set +a
```

### **Step 10**: Set up DB and apply migrations

create new database "comedian" in your local mysql agent
```
mysql -uroot -proot
create database comedian
```
From project root directory apply migrations with the following command: 
```
goose -dir "migrations" mysql "root:root@/comedian?parseTime=true" up
```

### **Step 10**: Start Comedian
Run: go run main.go

If your configuration is correct, you will receive a message from Comedian with simple "Hello Manager" text. Then proceed to check if slash commands are working. Try to get a list of comedian admins (/list admin) and see it it works. 

In case something does not work correctly double check the configuration and make sure you did not miss any installation steps.


## Usage

You can use following commands:
***
| Name | Hint | Description |
| --- | --- | --- | 
| /comedian | | displays helpful info about all commands |
| /comedian add | @user @user1 / (admin, pm, developer) | Adds a new user with selected role |
| /comedian remove | @user @user1 / (admin, pm, developer) | Removes user with selected role |
| /comedian show | (admin, pm, developer) | Shows users with selected role |
| /comedian add_deadline | hh:mm | Set standup time |
| /comedian show_deadline | - | Show standup time in current channel |
| /comedian remove_deadline | - | Delete standup time in current channel |
| /comedian add_timetable | @user1 @user2 on mon tue at 14:02 | Set individual standup time |
| /comedian show_timetable | @user1 @user2 | Show individual standup time for users |
| /comedian remove_timetable | @user1 @user2  | Delete individual standup time for users |
| /comedian report_on_project | #channelID 2017-01-01 2017-01-31 | gets all standups for specified project for time period |
| /comedian report_on_user | @user 2017-01-01 2017-01-31 | gets all standups for specified user for time period |
| /comedian report_on_user_in_project |@user #project 2017-01-01 2017-01-31 | gets all standups for specified user in project for time period |

***
First, you need to start remote standups meetings in Slack. 

Create a channel for it. Then add Comedian in the channel and ask your team to write messages with answers to the following questions tagging bot in the message:

1. What I did yesterday?
2. What I'm going to do today?
3. What problems I've faced?

Setup standups deadline when users should be ready to submit standups
Assign users to submit standups 

Enjoy automated remote standups meetings each morning! 

## Issues

Feel free to send pull requests. Also feel free to create issues.

## License

MIT License

Copyright (c) 2017 Mad Devs

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.