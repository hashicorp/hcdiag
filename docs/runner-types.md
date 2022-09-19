# Runner Types

hcdiag provides a variety of Runners for your Hashicorp-tool-troubleshooting convenience. Here is a table of currently available Runners:

| Constructor                | Config Block   | Description                                                                                                                                                                            | Parameters                                                          |
|----------------------------|----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------|
| `runner.NewCommander(...)` | `command`      | Issues a CLI command and optionally parses the result if format JSON is specified. Otherwise use string.                                                                               | `command = <string,required>` <br/> `format = <string,required>`    |
| `runner.NewCopier(...)`    | `copy`         | Copies the file or directory and all of its contents into the bundle using the same name. Since will check the last modified time of the file and ignore if it's outside the duration. | `path = <string,required>` <br/> `since = <duration,optional>`      |
| `runner.NewHTTPer(...)`    | `GET`          | Makes an HTTP get request to the path.                                                                                                                                                  | `path = <string,required>`                                          |
| `runner.NewSheller(...)`   | `shell`        | An "escape hatch" allowing arbitrary shell strings to be executed.                                                                                                                     | `run = <string,required>`                                           |
| `log.NewDocker(...)`       | `docker-log`   | Copies logs from a docker container, via the `docker logs` command.                                                                                                                    | `container = <string,required>` <br/> `since = <duration,optional>` | 
| `log.NewJournald(...)`     | `journald-log` | Copies logs from a journald service, via the `journalctl` command.                                                                                                                     | `service = <string,required>` <br/> `since = <duration,optional>`   |