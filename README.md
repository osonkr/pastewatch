# pastewatch

I wanted something to monitor pastebin leaks that match specific keywords.

## Usage
pastewatch configurations available:

| Flag         | ENV Variable            | Type   | Description                              | Default | Example                                             | Required |
|--------------|-------------------------|--------|------------------------------------------|---------|-----------------------------------------------------|----------|
| pastebin-key | PASTEBINKEY             | string | Pastebin Developer API Key               |         | PASTEBINKEY=xxxx --pastebin-key=xxxx                | Yes      |
| term         | TERMS (comma separated) | list   | Terms to watch new pastebins for         |         | TERMS=golang,pastebin --term golang --term pastebin | No       |
| interval     | REQUEST_INTERVAL        | int    | Interval to request bins at (in seconds) | 30      | REQUEST_INTERVAL=30 --interval 30                   | No       |
| limit        | REQUEST_LIMIT           | int    | Number of bins to retrieve per request   | 100     | REQUEST_LIMIT=100 --limit 100                       | No       |
| log-file     | LOG_FILE                | string | Path to log file, file will be created if one doesn't exist   |         | LOG_FILE=/var/log/pastewatch.log --log-file /var/log/pastewatch.log | No

#### From Docker:

`docker run daviddiefenderfer/pastewatch --pastebin-key=xxxxxxxxxx --term x --term y --term z`


#### From Kubernetes:

I use kubernetes to keep 1 instance running and store the search terms and key in a secret:

`kubectl create secret generic pastewatch-secrets --from-literal="PASTEBINKEY=xxxxx" --from-literal="TERMS=x,y,z"`

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: pastewatch
  name: pastewatch
spec:
  replicas: 1
  selector:
    matchLabels:
      run: pastewatch
  template:
    metadata:
      labels:
        run: pastewatch
    spec:
      volumes:
        - name: pastewatch-volume
          persistentVolumeClaim:
            claimName: pastewatch-vc
      containers:
      - image: daviddiefenderfer/pastewatch
        name: pastewatch
        resources: {}
        env:
          - name: LOG_FILE
            value: /var/log/pastewatch.txt
        envFrom:
          - secretRef:
              name: pastewatch-secrets
        volumeMounts:
          - name: pastewatch-volume
            mountPath: /var/log
```

## TODO:

Ideally I'd like for this to save findings to a database and use pub/sub service to send alerts

 - [ ] Allow findings to be saved to DB
 - [x] Write findings to optional logs file
 - [ ] Add drivers for other "pastebin"-like sites
   - [ ] Add support to run from proxy as these will likely be actually scraping pages

