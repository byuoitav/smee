import {Component, Inject, OnDestroy, OnInit} from '@angular/core';
import {MatTableDataSource} from "@angular/material/table";
import {MatDialog, MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";
import {Alert, ApiService, Issue, MaintenanceInfo} from "../api.service";

interface DialogData {
  roomID: string;
  room: string;
  maintenance: MaintenanceInfo;
}

@Component({
  selector: 'app-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit, OnDestroy {
  roomID: string = "";
  room: string = "";
  issue: Issue | undefined;
  maintenance: MaintenanceInfo | undefined;
  alertsDataSource: MatTableDataSource<Alert> = new MatTableDataSource(undefined);
  alertColumns: string[] = ["device", "type", "started", "ended"];
  updateInterval: number | undefined;

  constructor(private api: ApiService, private dialog: MatDialog) {}

  ngOnInit(): void {
    this.roomID = "ITB-1010"// get from route
    this.room = "ITB-1010"

    this.update();
    this.updateInterval = window.setInterval(() => {
      this.update();
    }, 10000);
  }

  ngOnDestroy(): void {
    if (this.updateInterval) {
      window.clearInterval(this.updateInterval);
    }
  }

  private update(): void {
    this.api.getIssue(this.roomID).subscribe(issue => {
      this.issue = issue;

      // TODO remove
      if (this.issue?.events) {
        this.issue.events.push(this.issue.events[0]);
      }

      if (this.issue?.alerts) {
        this.alertsDataSource = new MatTableDataSource([...this.issue.alerts.values()]);
      }
    });

    this.api.getMaintenanceInfo(this.roomID).subscribe(info => {
      this.maintenance = info;
    })
  }

  inMaintenance(): boolean {
    if (!this.maintenance?.start || !this.maintenance?.end) {
      return false;
    }

    const now = new Date();
    if (now < this.maintenance.start) {
      return false;
    } else if (now > this.maintenance.end) {
      return false;
    }

    return true;
  }

  editMaintenance(): void {
    const ref = this.dialog.open(MaintenanceDialog, {
      disableClose: true,
      data: {
        roomID: this.roomID,
        room: this.room,
        maintenance: this.maintenance,
      }
    });

    ref.afterClosed().subscribe(saved => {
      if (saved) {
        this.update();
      }
    })
  }
}

@Component({
  selector: 'app-maintenance-dialog',
  templateUrl: 'maintenance-dialog.html',
  styles: [
    `
    .content {
      display: flex;
      flex-direction: column;
    }
    `
  ],
})
export class MaintenanceDialog {
  info: MaintenanceInfo;

  constructor(private dialogRef: MatDialogRef<MaintenanceDialog, MaintenanceInfo>,
    private api: ApiService,
    @Inject(MAT_DIALOG_DATA) public data: DialogData) {
    this.info = {
      roomID: data.maintenance.roomID,
      start: data.maintenance.start,
      end: data.maintenance.end,
    };
  }

  parseDate(value: string): Date {
    return new Date(value);
  }

  canSave(): boolean {
    if (!this.info.start || !this.info.end) {
      return false;
    } else if (this.info.start > this.info.end) {
      return false;
    }

    // make sure this is in the future
    if (this.info.end < new Date()) {
      return false;
    }

    return true;
  }

  save(): void {
    if (!this.canSave()) {
      return;
    }

    this.api.setMaintenanceInfo(this.info).subscribe(info => {
      this.dialogRef.close(info);
    }, err => {
      console.log("unable to set maintenance info", err);
      // TODO show error popup
    })
  }

  // TODO
  disableMaintenance(): void {
  }
}
