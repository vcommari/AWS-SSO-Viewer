# AWS-SSO-Viewer

AWS-SSO-VIEWER allows you to quickly visualize all the AWS in your organization, see the permission set assigned to each account as well as get the permission sets details.

How to use :

- Export your AWS credentials (or use instance role)
- Setup config,yml with your SSO instanceARNm identity store id, port and aws region.
- Run go run main.go

Or with docker :

- Compile with make
- Setup config,yml with your SSO instanceARNm identity store id, port and aws region.
- Build the docker image
