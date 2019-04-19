
# awswrap

Execute command for all of aws environment.

## Usage

```
$ awswrap aws sts get-caller-identity
[default]       {
[default]           "UserId": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
[default]           "Account": "XXXXXXXXXXXXXXXXX",
[default]           "Arn": "arn:aws:sts::XXXXXXXXXXXXXXXXXXXXX:assumed-role/XXXXXXXXXXXXXXXXXXXXX/XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
[default]       }
[personal]      {
[personal]          "UserId": "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBbBBB",
[personal]          "Account": "YYYYYYYYYYYYYYYYY",
[personal]          "Arn": "arn:aws:sts::YYYYYYYYYYYYYYYYYYYY:assumed-role/YYYYYYYYYYYYYYYYYYYYYY/YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY"
[personal]      }
```

## How it works

Read all profiles in `~/.aws/credentials` and execute a command with environment provided AWS_PROFILE.

