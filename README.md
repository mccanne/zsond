# zsond

zsond is a simple test server for receiving zson over http.  It accepts zson
zeek file, looks for the #path directive, and writes the file to under current
directory of the running zsond using the path given in the POST endpoint naming
the log file and the path type.  If the file already exists, a different name
is used by embedding a version number in the file.

For example, run the server on port 9999:
```
git clone https://github.com/mccanne/zsond.git
cd zsond
go build
./zsond :9999
```
Then, point zeek's http-over-zson plugin at the server and run zeek.
You should see zeek logs appear in the server directory.

You can manually push a log into zsond with curl, e.g., to push conn.log:
```
curl -X POST "http://localhost:9999/foo" --data-binary @conn.log
```
You need --data-binary here as curl will otherwise strip newlines
from the log file.
