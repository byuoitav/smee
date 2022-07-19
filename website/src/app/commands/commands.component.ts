import { Component, OnInit } from '@angular/core';
import { CommandService } from 'app/command.service';


declare type actionCallback = (info: commandInfo) => void;

interface commandInfo{
  title: string,
  description: string,
  inputs: {label: string, value: string}[],
  actionButton: string,
  status: string,
  action: actionCallback
}
@Component({
  selector: 'app-commands',
  templateUrl: './commands.component.html',
  styleUrls: ['./commands.component.scss']
})
export class CommandsComponent implements OnInit {
  actionList: commandInfo[];

  constructor(private cS: CommandService) {
    this.actionList = [];

    this.registerActions();
  }

  ngOnInit(): void {
  }

  registerActions() {
    this.actionList = [
      {
        title: "Float",
        description: "Redeploy code to control pi",
        inputs: [
          {
            label: "Room or Device",
            value: ""
          }
        ],
        actionButton: "Float",
        status: "",
        action: (info: commandInfo) => {
          info.status = 'wait';
          if (info.inputs[0].value != "") {
            var resp = this.cS.float(info.inputs[0].value);
            resp.subscribe(
              data => {
                if (data == undefined) {
                  this.actionFailed(info);
                } else {
                  this.confirmAction(info);
                }
              },
              error => {});
          } else {
            this.missingInput(info);
          }
        }
      },
      {
        title: "Swab",
        description: "Refresh the local database on a control pi",
        inputs: [
          {
            label: "Room or Device",
            value: ""
          }
        ],
        actionButton: "Swab",
        status: "",
        action: (info: commandInfo) => {
          info.status = 'wait';
          if (info.inputs[0].value != "") {
            var resp = this.cS.swab(info.inputs[0].value);
            resp.subscribe(
              data => {
                if (data == undefined) {
                  this.actionFailed(info);
                } else {
                  this.confirmAction(info);
                }
              },
              error => {});
          } else {
            this.missingInput(info);
          }
        }
      },
      {
        title: "Sink",
        description: "Reboot control pis",
        inputs: [
          {
            label: "Room or Device",
            value: ""
          }
        ],
        actionButton: "Sink",
        status: "",
        action: (info: commandInfo) => {
          info.status = 'wait';
          if (info.inputs[0].value != "") {
            var resp = this.cS.sink(info.inputs[0].value);
            resp.subscribe(
              data => {
                if (data == undefined) {
                  this.actionFailed(info);
                } else {
                  this.confirmAction(info);
                }
              },
              error => {});
          } else {
            this.missingInput(info);
          }
        }
      },
      {
        title: "Fix Time",
        description: "Sync time on control pis with BYU's servers",
        inputs: [
          {
            label: "Room or Device",
            value: ""
          }
        ],
        actionButton: "Fix",
        status: "",
        action: (info: commandInfo) => {
          info.status = 'wait';
          if (info.inputs[0].value != "") {
            var resp = this.cS.fixTime(info.inputs[0].value);
            resp.subscribe(
              data => {
                if (data == undefined) {
                  this.actionFailed(info);
                } else {
                  this.confirmAction(info);
                }
              },
              error => {});
          } else {
            this.missingInput(info);
          }
        }
      },
      {
        title: "Remove Device",
        description: "Remove a device from monitoring. This does NOT remove the device from the database",
        inputs: [
          {
            label: "Device",
            value: ""
          }
        ],
        actionButton: "Remove",
        status: "",
        action: (info: commandInfo) => {
          info.status = 'wait';
          if (info.inputs[0].value != "") {
            var resp = this.cS.rmDevice(info.inputs[0].value);
            resp.subscribe(
              data => {
                if (data == undefined) {
                  this.actionFailed(info);
                } else {
                  this.confirmAction(info);
                }
              },
              error => {});
          } else {
            this.missingInput(info);
          }
        }
      },
      {
        title: "Close Issue",
        description: "Close an issue in monitoring",
        inputs: [
          {
            label: "Room",
            value: ""
          }
        ],
        actionButton: "Close",
        status: "",
        action: (info: commandInfo) => {
          info.status = 'wait';
          if (info.inputs[0].value != "") {
            var resp = this.cS.closeIssue(info.inputs[0].value);
            resp.subscribe(
              data => {
                if (data == undefined) {
                  this.actionFailed(info);
                } else {
                  this.confirmAction(info);
                }
              },
              error => {});
          } else {
            this.missingInput(info);
          }
        }
      },
      {
        title: "Duplicate Database",
        description: "Duplicate a room within the database",
        inputs: [
          {
            label: "Source Device",
            value: ""
          },
          {
            label: "Destination Device",
            value: ""
          }
        ],
        actionButton: "Duplicate",
        status: "",
        action: (info: commandInfo) => {
          info.status = 'wait';
          if (info.inputs[0].value != "" && info.inputs[1].value != "") {
            var resp = this.cS.dupDatabase(info.inputs[0].value, info.inputs[1].value);
            resp.subscribe(
              data => {
                if (data == undefined) {
                  this.actionFailed(info);
                } else {
                  this.confirmAction(info);
                }
              },
              error => {});
          } else {
            this.missingInput(info);
          }
        }
      }
    ]
  }

  trackByFn(index: any, item: any) {
    return index;
  }

  missingInput(cmdInfo: commandInfo) {
    cmdInfo.status = "fail";
    this.timeoutStatus(cmdInfo);
  }

  confirmAction(cmdInfo: commandInfo) {
    cmdInfo.status = "success";
    this.timeoutStatus(cmdInfo);
  }

  actionFailed(cmdInfo: commandInfo) {
    cmdInfo.status = "fail"
    this.timeoutStatus(cmdInfo);
  }

  timeoutStatus(info: commandInfo) {
    setTimeout(() => {info.status = "";}, 5000);
  }
}
