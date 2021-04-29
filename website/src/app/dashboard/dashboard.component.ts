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

    ref.afterClosed().subscribe(issue => {
      if (!issue || !issue.incidents) {
        this.updateIssues();
        return;
      }

      for (const [_, inc] of issue.incidents) {
        window.open(this.incidentLink(inc));
        break;
      }

      this.updateIssues();
    })
  }

  linkIncident(issue: Issue): void {
    const ref = this.dialog.open(DashboardLinkDialog, {
      data: {
        issue: issue,
      }
    });

    ref.afterClosed().subscribe(issue => {
      if (!issue || !issue.incidents) {
        this.updateIssues();
        return;
      }

      for (const [_, inc] of issue.incidents) {
        window.open(this.incidentLink(inc));
        break;
      }

      this.updateIssues();
    })
  }

  incidentLink(inc: Incident): string {
    return `https://support.byu.edu/nav_to.do?uri=task.do?sys_id=${inc.id}`;
  }
}

@Component({
  selector: 'app-dashboard-create-dialog',
  templateUrl: 'create-dialog.html',
})
export class DashboardCreateDialog {
  shortDesc: string | undefined = undefined;

  constructor(private dialogRef: MatDialogRef<DashboardLinkDialog, Issue>,
    private api: ApiService,
    @Inject(MAT_DIALOG_DATA) public data: DialogData) {}

  create(): void {
    if (!this.shortDesc) {
      return;
    }

    this.api.createIncidentFromIssue(this.data.issue.id, this.shortDesc).subscribe(issue => {
      this.dialogRef.close(issue);
    }, err => {
      console.log("unable to link issue", err);

      // TODO show error popup
    });
  }
}

@Component({
  selector: 'app-dashboard-link-dialog',
  templateUrl: 'link-dialog.html',
})
export class DashboardLinkDialog {
  incidentName: string | undefined = undefined;

  constructor(private dialogRef: MatDialogRef<DashboardLinkDialog, Issue>,
    private api: ApiService,
    @Inject(MAT_DIALOG_DATA) public data: DialogData) {}

  link(): void {
    if (!this.incidentName) {
      return;
    }

    this.api.linkIssueToIncident(this.data.issue.id, this.incidentName).subscribe(issue => {
      this.dialogRef.close(issue);
    }, err => {
      console.log("unable to link issue", err);

      // TODO show error popup
    });
  }
}
