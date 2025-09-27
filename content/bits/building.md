---
Title: Building a Static Site Generator
DatePublished: 09/18/2025
---
## Why build your own over any of the 
The world wide web is filled with trillions of web pages
- many of them are bloated with ads and offer a terrible user experience
- some of them were written in the 90s and 2000's and still exist today, more performant than most modern stacks
I've deployed a bunch of sites using Rails, PostgreSQL, and Heroku. It works perfectly well but it's overkill for a portfolio. The inspiration for this project was coming across [brandur.org](https://brandur.org/). He has a short piece on why he moved from a rails stack to a custom Go executable and I resonated with his reasons and they pointed in the same direction as my observations about the early web. 
### Static sites fit well into the mold of minimalism
only take up the resources necessary to run the site. Navigating the internet in the year 2025 is far from ideal. Annoying ads, logins for every site, unnecessary popups. Incredibly long load times. The list is unexhaustible. 
 
Bullet points should be in order from most to least important
- I'm currently learning Golang so this works well as a project to get familiar with the language.
- ease of hosting
- price of hosting
- Maintainability
- Speed

The stack, at moment of writing this, is two S3 buckets (storage service from Amazon)
1. root domain - harrisonblake.net
2. subdomain - www.harrisonblake.net

All content is kept in the subdomain S3 bucket and any request to the root will redirect to the subdomain. The repo is stored on Github and we use github actions to automate deployment. The entire stack is provided by Amazon and it makes hosting painless.
### AWS Prerequisites

**AWS account**
On account creation you have a root user that has full access to all AWS resources
- It is recommended to create an IAM user and grant specific permissions for security

**IAM user**
Creates and gives access / permission to resources set by root user. DO NOT GET CONFUSED WITH IAM IDENTITY. IAM and IAM IDENTITY are 2 separate ways to achieve the same *thing*(allow certain access to certain entities for security).
1. create admin group and add new user to group
2. give group admin permissions
3. enable console (creates password)
4. log into IAM user
	AWS Account ID is needed to login (goto any IAM page and click on the username in top right corner. This open a drop down menu revealing the Account ID)
5. create an access key
	in the same drop down menu from the last step, click on `Security credentials`

**AWS CLI**
install cli tool with the following commands
```
curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
sudo installer -pkg AWSCLIV2.pkg -target /
```
verify install with command
	`which aws`
configure with command
	`aws configure`
	![[zConfigureAWS.png]]

**Register a Domain**
I'm going with route 53. Why stray from the all AWS stack? 

**Configure a Certificate using ACM(AWS Certificate Manager)**
Make sure to specify...
1. root domain `example.com`
2. subdomain(s) `*.example.com`

**Verify you own the domain name**
Route 53 will auto verify with the click of a button (the all Amazon stack has its perks).

**S3 Buckets**

**Cloudfront Distribution**

**File Structure**

**CI/CD w/ Github actions**
