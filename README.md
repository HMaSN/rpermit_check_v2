# rpermit_check_v2

This app checks the status of the German residence permit.

The app uses Redis, the permit ID as the key and the last status as value.

Mailgun is used to send email notification for the updates.

| Constant | Description |
| ------ | ------ |
|CheckURL|the URL used to check the issuing of the Permit status|
|RedisURL|Redis URL|
|RedisUser|Redis instance user|
|RedisPassword|Redis instance password|
|zapNummer|the Permit request ID|
|MailGunURL|API URL for mailgun|
|MailGunAPIToken|api token for Mailgun|

