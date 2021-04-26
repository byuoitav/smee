import {Component, Inject, OnDestroy, OnInit} from '@angular/core';
import {MatDialog, MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";
import {MatTableDataSource} from "@angular/material/table";
import {ApiService, Issue, Alert, Incident, IssueEvent} from "../api.service";

interface DialogData {
  issue: Issue;
}

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit, OnDestroy {
  displayedColumns: string[] = ["room", "alertCount", "started", "incidents"];
  dataSource: MatTableDataSource<Issue> = new MatTableDataSource(undefined);
  issueUpdateInterval: number | undefined;

  constructor(private api: ApiService, private dialog: MatDialog) {}

  ngOnInit(): void {
    this.updateIssues();
    this.issueUpdateInterval = window.setInterval(() => {
      this.updateIssues();
    }, 10000);
  }

  ngOnDestroy(): void {
    if (this.issueUpdateInterval) {
      window.clearInterval(this.issueUpdateInterval);
    }
  }

  private updateIssues(): void {
    this.api.getIssues().subscribe(issues => {
      this.dataSource = new MatTableDataSource(issues);
    })
  }

  createIncident(issue: Issue): void {
    const ref = this.dialog.open(DashboardCreateDialog, {
      data: {
        issue: issue,
      }
    });

    ref.afterClosed().subscribe(() => {
    })
  }

  linkIncident(issue: Issue): void {
    const ref = this.dialog.open(DashboardLinkDialog, {
      data: {
        issue: issue,
      }
    });

    ref.afterClosed().subscribe(() => {
    })
  }
}

@Component({
  selector: 'app-dashboard-create-dialog',
  templateUrl: 'create-dialog.html',
})
export class DashboardCreateDialog {
  constructor(@Inject(MAT_DIALOG_DATA) public data: DialogData) {}
}

@Component({
  selector: 'app-dashboard-link-dialog',
  templateUrl: 'link-dialog.html',
})
export class DashboardLinkDialog {
  incidentName: string | undefined = undefined;

  constructor(private dialogRef: MatDialogRef<DashboardLinkDialog>,
    private api: ApiService,
    @Inject(MAT_DIALOG_DATA) public data: DialogData) {}

  link(): void {
    if (!this.incidentName) {
      return;
    }

    this.api.linkIssueToIncident(this.data.issue.id, this.incidentName).subscribe(issue => {
      this.dialogRef.close();
    }, err => {
      console.log("unable to link issue", err);
      // TODO show error popup
    });
  }
}
