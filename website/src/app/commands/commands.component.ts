import { Component, OnInit } from '@angular/core';
import { CommandService } from 'app/command.service';


declare type actionCallback = (info: commandInfo) => void;

interface commandInfo{
  title: string,
  description: string,
  label: string[],
  actionButton: string,
  input: string[],
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
        label: ["Room or Device"],
        actionButton: "Float",
        input: [""],
        status: "",
        action: (info: commandInfo) => {
          if (info.input[0] != "") {
            var resp = this.cS.float(info.input[0]);
            if (resp) {} else {}
          }
          this.missingInput(info);
        }
      },
      {
        title: "Swab",
        description: "Refresh data in the database on control pis",
        label: ["Room or Device"],
        actionButton: "Swab",
        input: [""],
        status: "",
        action: (info: commandInfo) => {
          if (info.input[0] != "") {
            var resp = this.cS.swab(info.input[0]);
            if (resp) {} else {}
          }
          this.missingInput(info);
        }
      },
      {
        title: "Sink",
        description: "Reboot control pis",
        label: ["Room or Device"],
        actionButton: "Sink",
        input: [""],
        status: "",
        action: (info: commandInfo) => {
          if (info.input[0] != "") {
            var resp = this.cS.sink(info.input[0]);
            if (resp) {} else {}
          }
          this.missingInput(info);
        }
      },
      {
        title: "Fix Time",
        description: "Sync time on control pis with BYU's servers",
        label: ["Room or Device"],
        actionButton: "Fix",
        input: [""],
        status: "",
        action: (info: commandInfo) => {
          if (info.input[0] != "") {
            var resp = this.cS.fixTime(info.input[0]);
            if (resp) {} else {}
          }
          this.missingInput(info);
        }
      },
      {
        title: "Remove Device",
        description: "Remove a device from monitoring. This does NOT remove the device from the database",
        label: ["Device"],
        actionButton: "Remove",
        input: [""],
        status: "",
        action: (info: commandInfo) => {
          if (info.input[0] != "") {
            var resp = this.cS.rmDevice(info.input[0]);
            if (resp) {} else {}
          }
          this.missingInput(info);
        }
      },
      {
        title: "Close Issue",
        description: "Close an issue in monitoring",
        label: ["Room"],
        actionButton: "Close",
        input: [""],
        status: "",
        action: (info: commandInfo) => {
          if (info.input[0] != "") {
            var resp = this.cS.closeIssue(info.input[0]);
            if (resp) {} else {}
          }
          this.missingInput(info);
        }
      },
      {
        title: "Duplicate Database",
        description: "Duplicate a room within the database",
        label: ["Source Device", "Destination Device"],
        actionButton: "Duplicate",
        input: ["", ""],
        status: "",
        action: (info: commandInfo) => {
          if (info.input[0] != "" && info.input[1] != "") {
            var resp = this.cS.dupeDatabase(info.input[0], info.input[1]);
            if (resp) {} else {}
          }
          this.missingInput(info);
        }
      },
      {
        title: "Screenshot",
        description: "Take and display a screenshot of a control pi",
        label: ["Device"],
        actionButton: "Screenshot",
        input: [""],
        status: "",
        action: (info: commandInfo) => {} // use service to hit endpoint, display resulting image
      }
    ]
  }

  trackByFn(index: any, item: any) {
    return index;
  }
  
  missingInput(disp: commandInfo) {
    disp.status = "Cannot complete action. Please provide proper input.";
    this.timeoutStatus(disp);
  }

  timeoutStatus(info: commandInfo) {
    setTimeout(() => {info.status = "";}, 10000);
  }
}
