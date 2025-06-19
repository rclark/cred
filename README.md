# cred

Helps you set AWS credentials as explicit environment variables in your terminal.

### Use

These examples will set the following environment variables:

- `AWS_ACCOUNT_ID`
- `AWS_DEFAULT_REGION`
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_SESSION_TOKEN` (if applicable)
- `AWS_SESSION_EXPIRES_AT` (if applicable)

```sh
> eval $(cred)
> echo $AWS_ACCESS_KEY_ID
ASIA5M45ETFUAF7H4XFY
```

or

```sh
> eval $(cred --profile my-profile)
> echo $AWS_ACCESS_KEY_ID
ASIA3K82BVNR9P6M2TDL
```

### Notes

- You must first set up a proper `~/.aws/config` file for yourself.

- Generally, if you're using a program that relies on an AWS SDK, you shouldn't have to use this tool. The program should handle getting credentials for you, using your `~/.aws/config` file, whenever it needs them. Use this only in a situation where you **must** have explicit credentials in your environment.

- Keep in mind that most decent `~/.aws/config` profiles will give a set of _temporary_ credentials. You can run `cred expiry` to print the time when the credentials set in your environment will expire.
