## Usage example

``` sh
crontab -l | ./insert-cronjob-into-crontab --alias "test" --cronjob "0 1 * * mon /bin/true"
```
