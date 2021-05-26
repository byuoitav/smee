import {Component, Inject, OnDestroy, OnInit, ViewChild, AfterViewInit} from '@angular/core';
import {MatDialog, MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";
import {MatPaginator} from "@angular/material/paginator";
import {MatSort} from "@angular/material/sort";
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
export class DashboardComponent implements OnInit, OnDestroy, AfterViewInit {
  displayedColumns: string[] = ["room", "alertCount", "alertOverview", "age", "incidents"];
  dataSource: MatTableDataSource<Issue> = new MatTableDataSource(undefined);
  issueUpdateInterval: number | undefined;
  showMaintenance: boolean = false;

  @ViewChild(MatPaginator) paginator: MatPaginator | null = null;
  @ViewChild(MatSort) sort: MatSort | null = null;

  constructor(private api: ApiService, private dialog: MatDialog) {}

  ngOnInit(): void {
    this.updateIssues();
    this.issueUpdateInterval = window.setInterval(() => {
      this.updateIssues();
    }, 10000);

    this.dataSource.filterPredicate = (data: Issue, filter: string): boolean => {
      const dataList = [];
      dataList.push(data.room.id.toLowerCase());
      dataList.push(data.room.name.toLowerCase());

      const dataStr = dataList.join("◬");

      const transformedFilter = filter.trim().toLowerCase();
      return dataStr.includes(transformedFilter)
    };

    this.dataSource.sortData = (data: Issue[], sort: MatSort): Issue[] => {
      if (!sort.active || sort.direction === '') {
        return data;
      }

      const isAsc = sort.direction === 'asc';

      const cmp = (a: number | string | Date | undefined, b: number | string | Date | undefined): number => {
        // return -1 if a is less than b
        // return 1 if b is less than a
        if (!a && !b) {
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
          case 'room': return cmp(a.room.name, b.room.name);
          case 'alertCount': return cmp(a.alerts?.size, b.alerts?.size);
          case 'age': return cmp(a.start, b.start);
          default: return 0;
        }
      });
    }
  }

  ngAfterViewInit() {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
  }

  ngOnDestroy(): void {
    if (this.issueUpdateInterval) {
      window.clearInterval(this.issueUpdateInterval);
    }
  }

  private updateIssues(): void {
    this.api.getIssues().subscribe(issues => {
      this.dataSource.data = issues;
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

  applyFilter(event: Event) {
    const filterValue = (event.target as HTMLInputElement).value;
    this.dataSource.filter = filterValue.trim().toLowerCase();
  }

  alertOverview(issue: Issue): string {
    if (!issue.alerts) {
      return "No alerts";
    }

    // map of type -> array of names that have that type
    const alerts = new Map<string, string[]>();
    for (const [id, alert] of issue?.alerts.entries()) {
      if (alert.end) {
        // skip inactive alerts
        // TODO fix backend to not send these
        console.log("skipping inactive alert");
        continue;
      }

      const split = alert.device.name.split("-");
      let name = alert.device.name;
      if (split.length == 3) {
        name = split[2];
      }

      if (alerts.has(alert.type)) {
        alerts.get(alert.type)?.push(name);
      } else {
        alerts.set(alert.type, [name]);
      }
    }

    const groups: string[] = [];
    for (const [type, devices] of alerts) {
      groups.push(type + ` (${devices.join(", ")})`);
    }

    return groups.join(", ");
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
