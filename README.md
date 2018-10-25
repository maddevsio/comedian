<div align="center">
    <img style="width: 300px" src ="documentation/logo.png" />
</div>

<div align="center"> Stand up bot for slack to help you with remote stand up meetings automation </div>

<div align="center">

![](https://travis-ci.org/maddevsio/comedian.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/maddevsio/comedian/badge.svg)](https://coveralls.io/github/maddevsio/comedian)
[![Go Report Card](https://goreportcard.com/badge/github.com/maddevsio/comedian)](https://goreportcard.com/report/github.com/maddevsio/comedian)
[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)
[![](https://godoc.org/github.com/maddevsio/comedian?status.svg)](https://godoc.org/github.com/maddevsio/comedian)

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
- [x] Provide daily report on team's yesterday performance, weekly report on Sundays
- [x] Support English and Russian languages


## Getting started locally

These instructions will help you set up the project on your local machine for development and testing purposes with [ngrok](https://ngrok.com/product) 


### **Step 1**: Create a public HTTPS URL for Comedian
Install [ngrok](https://ngrok.com/product) and create a public HTTPS URL for Comedian on your development machine following the instruction from the web

### **Step 2**: Clone the project
Copy the project repository to your local machine. Note: Go should be already installed! If you do not have Go installed, please, follow [installation guidelines](https://golang.org/doc/install) from Go official website to install it and then proceed to Step 2

```
mkdir -p $GOPATH/src/github.com/maddevsio/
cd $GOPATH/src/github.com/maddevsio/
git clone https://github.com/maddevsio/comedian
```

### **Step 3**: Configure environmental variables

Create `.env` file in the root directory and add the following env variables there. See .env.example for a reference:

| Title | Description | Default | Optional? |
| --- | --- | --- | --- |
| COMEDIAN_SLACK_TOKEN | Bot User OAuth Access Token |  | No |
| COMEDIAN_DATABASE | Database URL. Default: comedian:comedian@/comedian?parseTime=true |  | No |
| COMEDIAN_SECRET_TOKEN | Include to secure Comedian API |  | Yes |
| COMEDIAN_HTTP_BIND_ADDR | HTTP bind address | 0.0.0.0:8080 | No |
| COMEDIAN_LANGUAGE | Comedian primary language | en_US | No |
| COMEDIAN_SUPER_ADMIN_ID | Slack ID of super admin in your workspace |  | No |
| COMEDIAN_REPORT_CHANNEL | Slack Channel ID to send daily reports to |  | No |
| COMEDIAN_REPORT_TIME | Time to send daily reports | 10:00 | No |
| COMEDIAN_MAX_REMINDERS | Number of times comedian keeps reminding non reporters | 3 | No |
| COMEDIAN_REMINDER_INTERVAL | Duration of the intervals when Comedian waits before next reminder in minutes | 30 | No |
| COMEDIAN_WARNING_TIME | Duration prior to deadline to remind about upcoming deadline | 10 | No |
| COMEDIAN_ENABLE_COLLECTOR | Enables or Disables Collector* API requests | false | Yes |
| COMEDIAN_COLLECTOR_TOKEN | Secret Token for Collector* API requests |  | Yes |
| COMEDIAN_COLLECTOR_URL | URL to send Collector* API requests |  | Yes |
| COMEDIAN_SLACK_DOMAIN | Slack workspace title (copy first word of the link) |  | Yes |
| TZ | Setup time zone for comedian DB | UTC | Yes |

*Please note that Collector Servise is developed only for internal use of Mad Devs LLC, therefore when configuring Comedian, you may turn this feature off.

### **Step 4**: Create Slack chatbot 
Create "app" in slack workspace: https://api.slack.com/apps
In the drop-down list at the top select the created "app"

### **Step 5**: Configure slash commands
In the menu, select "Slash Commands". Create the following commands (Request URL for all commands: ```http://<ngrok https URL>/commands(here you can paste COMEDIAN_SECRET_TOKEN if it is not empty) ``` )

| Name | Hint | Description | Escape option |
| --- | --- | --- | --- |
| /helper | | displays helpful info about slash commands | - |
| /add | @user @user1 / (admin, pm, developer) | Adds a new user with selected role | V |
| /delete | @user @user1 / (admin, pm, developer) | Removes user with selected role  | V |
| /list | (admin, pm, developer) | Lists users with selected role | - |
| /standup_time_set | hh:mm | Set standup time | - |
| /standup_time | - | Show standup time in current channel | - |
| /standup_time_remove | - | Delete standup time in current channel | - |
| /timetable_set | @user1 @user2 on mon tue at 14:02 | Set individual standup time | V |
| /timetable_show | @user1 @user2 | Show individual standup time for users | V |
| /timetable_remove | @user1 @user2  | Delete individual standup time for users | V |
| /report_by_project | #channelID 2017-01-01 2017-01-31 | gets all standups for specified project for time period | - |
| /report_by_user | @user 2017-01-01 2017-01-31 | gets all standups for specified user for time period | - |
| /report_by_user_in_project | #project @user 2017-01-01 2017-01-31 | gets all standups for specified user in project for time period | - |

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

### **Step 10**: Start Comedian
Run: go run main.go

If your configuration is correct, you will receive a message from Comedian with simple "Hello Manager" text. Then proceed to check if slash commands are working. Try to get a list of comedian admins (/list admin) and see it it works. 

In case something does not work correctly double check the configuration and make sure you did not miss any installation steps.


## Deploy on [Digital Ocean](https://www.digitalocean.com/pricing/)
If you are willing to use Comedian for your organization, we recommend you to proceed with Digital Ocean droplet. Here is the basic instructions how to deploy Comedian to DO:

### **Step 1**: Purchase Droplet
1. Login to Digital Ocean
2. Add new project 
3. Add a droplet, choose Ubuntu 18.10 as your distributive
4. Select $5 per month plan
5. Do not add ssh key (if you are new to configuration)
6. Choose a name for your droplet
7. Press "create" button

After some time you will get an email with all info needed to login to your droplet

### **Step 2**: Configure the server
Login to your newly created server using SSH. You should have recieved email from Digital Ocean with IP address, login and a password. 

Open your terminal, type `ssh login@ipaddress` and then insert password. If this is your first time using this server, DO will ask you to reset the password. 

Next step is to [install docker-compose](https://www.digitalocean.com/community/tutorials/how-to-install-docker-compose-on-ubuntu-18-04) and [docker](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-on-ubuntu-18-04)


### **Step 3**: Prepare docker-compose file 

run `nano docker-compose.yml` to create your docker-compose file. Use docker-compose.yml file from the repository to set up it properly. You can use env variables inside or just type parameters right there for more readability. 

Make sure to use updated Comedian images from [DockerHub](HTTPS://hub.docker.com/r/anatoliyfedorenko/comedian/tags/) 

Save the changes and proceed to the next step. If you use .env file, make sure you updated your env variables before Step 4.

If you care about security, you may use HTTPS://github.com/JrCs/docker-letsencrypt-nginx-proxy-companion to configure Let's Encrypt certificates. 


### **Step 4**: Install chatbot to your workspace
Follow Steps 4-9 of the local installation guidelines before you proceed! 

### **Step 5**: Use docker-compose 

Once your docker-compose.yml file is ready and you installed Comedian App in your workspace, run `docker-compose up` or `docker-compose up -d` to start Comedian on the background.

Follow Step 10 of the local installation guide to check if Comedian was installed successfully. 

## Usage

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