import {AfterViewInit, Component, Inject, OnDestroy, OnInit} from '@angular/core';
import {MatTableDataSource} from "@angular/material/table";
import {MatSort} from "@angular/material/sort";
import {MatDialog, MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";
import {Alert, ApiService, Issue, MaintenanceInfo, IssueType, IssueTypeMap} from "../api.service";
import {ActivatedRoute} from "@angular/router";
import { ViewChild } from '@angular/core';

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
export class RoomComponent implements OnInit, OnDestroy, AfterViewInit {
  alertColumns: string[] = ["device", "type", "start", "end", "serviceNow"];
  updateInterval: number | undefined;
  alertsDataSource: MatTableDataSource<Alert> = new MatTableDataSource(undefined);
  roomID: string = "";
  roomName: string = "";
  issue: Issue | undefined;
  maintenance: MaintenanceInfo | undefined;
  issueType : IssueTypeMap | undefined;

  @ViewChild(MatSort) sort: MatSort | null = null;
  
  constructor(private api: ApiService, private dialog: MatDialog, private route: ActivatedRoute) {}

  ngOnInit(): void {
    this.alertsDataSource.sort = this.sort;
    this.api.getIssueType().subscribe(info =>{
      this.issueType = info;
    });
    this.update();
    this.route.params.subscribe(params => {
      this.roomID = params["roomID"];
      this.roomName = this.roomID; // TODO get from update()
      this.update();
    });

    this.updateInterval = window.setInterval(() => {
      this.update();
    }, 10000);
    
    this.alertsDataSource.sortData = (data: Alert[], sort: MatSort): Alert[] => {
      if (!sort.active || sort.direction === ''){
        return data;
      }
      const isAsc = sort.direction === 'asc';

      const cmp = (a: string | Date | undefined, b: string | Date | undefined): number => {
        if (!a && !b){
          return 0;
        } else if (!b) {
          return -1;
        } else if (!a) {
          return 1;
        }

        return (a < b ? -1 : 1) * (isAsc ? 1 : -1);
      }

      return data.sort((a, b) => {
        switch (sort.active) {
          case 'type': return cmp (a.type, b.type);
          case 'start': return cmp (a.start, b.start);
          case 'end': return cmp (a.end, b.end);
          default: return 0;
        }
      });
    }

  }
  ngAfterViewInit() {
    this.alertsDataSource.sort = this.sort;
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
        this.alertsDataSource.sort = this.sort;
      }
      this.api.getMaintenanceInfo(this.roomID).subscribe(info => {
        this.maintenance = info;
      })
    });
  }

  IssueTypeUrl(alert : Alert): string{
    const IssueMap = this.issueType?.IssueType
    const url = "https://it.byu.edu/nav_to.do?uri=kb_view.do?sysparm_article="
    var issuetypeurl = url + IssueMap?.get(alert.type)?.kbArticle
    return issuetypeurl
  }

  isInIssueType(alert : Alert): boolean{
    const IssueMap = this.issueType?.IssueType
    if (!IssueMap){
      return false
    }
    return IssueMap.has(alert.type)
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
      note: data.maintenance.note ? data.maintenance.note : undefined
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
    this.info.note = undefined;
    
    this.api.setMaintenanceInfo(this.info).subscribe(info => {
      this.dialogRef.close(info);
    }, err => {
      // TODO show error popup
      console.log("unable to disable maintenance", err);
    })
  }
}
