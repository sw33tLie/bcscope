# bcscope
Get the scope of your bugcrowd programs!

Example command:
```
go run main.go -t <YOUR-TOKEN-HERE> -c 2 -p
```
This will print all the scope of all your Bugcrowd private programs.
Remove the -p flag to get public programs too.
Keep the concurrency low otherwise arcwhite may not be so happy :)

```
go run main.go -h

Arguments:

  -h  --help         Print help information
  -t  --token        Bugcrowd session token (_crowdcontrol_session)
  -p  --private      Only show private invites. Default: false
  -l  --list         List programs instead of grabbing their scope. Default:
                     false
  -b  --bbp          Only show programs offering monetary rewards. Default:
                     false
  -c  --concurrency  Set concurrency. Default: 2
  -u  --url          Also print the program URL. Default: false

```
