import {Component, Inject, OnDestroy, OnInit} from '@angular/core';
import {MatTableDataSource} from "@angular/material/table";
import {MatDialog, MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";
import {Alert, ApiService, Issue, MaintenanceInfo} from "../api.service";
import {ActivatedRoute} from "@angular/router";

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
  alertColumns: string[] = ["device", "type", "started", "ended"];
  updateInterval: number | undefined;
  alertsDataSource: MatTableDataSource<Alert> = new MatTableDataSource(undefined);
  
  roomID: string = "";
  roomName: string = "";
  issue: Issue | undefined;
  maintenance: MaintenanceInfo | undefined;

  constructor(private api: ApiService, private dialog: MatDialog, private route: ActivatedRoute) {}

  ngOnInit(): void {
    this.route.params.subscribe(params => {
      this.roomID = params["roomID"];
      this.roomName = this.roomID; // TODO get from update()

      this.update();
    })

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
    if (!this.roomID) {
      return;
    }

    this.api.getIssue(this.roomID).subscribe(issue => {
      this.issue = issue;

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
        room: this.roomName,
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
      roomID: data.roomID,
      start: data.maintenance.start ? data.maintenance.start : new Date(),
      end: data.maintenance.end ? data.maintenance.end : new Date(new Date().getTime() + 60 * 60 * 24 * 1000),
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

  disable(): void {
    this.info.start = undefined;
    this.info.end = undefined;

    this.api.setMaintenanceInfo(this.info).subscribe(info => {
      this.dialogRef.close(info);
    }, err => {
      // TODO show error popup
      console.log("unable to disable maintenance", err);
    })
  }
}
