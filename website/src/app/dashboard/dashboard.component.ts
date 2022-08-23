import {Component, Inject, OnDestroy, OnInit, ViewChild, AfterViewInit} from '@angular/core';
import {MatDialog, MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";
import {MatPaginator} from "@angular/material/paginator";
import {MatSort} from "@angular/material/sort";
import {MatTableDataSource} from "@angular/material/table";
import {ApiService, Issue, Incident} from "../api.service";

interface DialogData {
  issue: Issue;
}

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit, OnDestroy, AfterViewInit {
  displayedColumns: string[] = ["room", "maintenance","alertCount", "alertOverview", "age", "incidents", "status"];
  unacknowledgedColumns: string[] = ["room", "maintenance", "alertCount", "alertOverview", "age", "acknowledge"];
  issueUpdateInterval: number | undefined;
  dataSource: MatTableDataSource<Issue> = new MatTableDataSource(undefined);
  unacknowledgedDataSource: MatTableDataSource<Issue> = new MatTableDataSource(undefined);
  showMaintenance: boolean = true;
  filterValue: string = "";
  totalAlerts: number = 0;
  totalIssues: number = 0;

  @ViewChild('akwPaginator') akwPaginator: MatPaginator | null = null;
  @ViewChild('akwTable', {read: MatSort, static: true}) akwSort: MatSort | null = null;


  @ViewChild('unakwPaginator') unakwPaginator: MatPaginator | null = null;
  @ViewChild('unakwTable', {read: MatSort, static: true}) unakwSort: MatSort | null = null;


  constructor(private api: ApiService, private dialog: MatDialog) {}

  ngOnInit(): void {
    this.updateIssues();
    
    this.issueUpdateInterval = window.setInterval(() => {
      this.updateIssues();
    }, 10000);

    var filterPredicate = (data: Issue, filter: string): boolean => {      
      if(!this.showMaintenance){
        if(data.isOnMaintenance){
          
          return false;
        }
      }
      //This is a workaround: The filter predicate does not run on an empty string.
      //This unicode replaces the empty filter value (it is not likely that the user would use a unicode character in the search filter)
      if(filter === "◬"){
        return true;
      }
      //--------------------------------//

      const dataList = [];
      dataList.push(data.room.id.toLowerCase());
      dataList.push(data.room.name.toLowerCase());

      if (data.alerts) {
        for (const [_, v] of data.alerts) {
          dataList.push(v.device.id.toLowerCase());
          dataList.push(v.device.name.toLowerCase());
          dataList.push(v.type.toLowerCase());
        }
      }
      const dataStr = dataList.join("◬");

      if (filter.charAt(0) == '-'){
        var newfilter = filter.substring(1);
        const transformedFilter = newfilter.trim().toLowerCase();
        return !dataStr.includes(transformedFilter)

      }else{
        const transformedFilter = filter.trim().toLowerCase();
        return dataStr.includes(transformedFilter)
      }

    };

    this.dataSource.filterPredicate = filterPredicate;
    this.unacknowledgedDataSource.filterPredicate = filterPredicate;

    var sortConfig = (data: Issue[], matSort: MatSort): Issue[] => {
      if (!matSort.active || matSort.direction === '') {
        return data;
      }

      const isAsc = matSort.direction === 'asc';

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
        switch (matSort.active) {
          case 'room': return cmp(a.room.name, b.room.name);
          case 'alertCount': return cmp(this.getActiveAlerts(a), this.getActiveAlerts(b));
          case 'age': return cmp(a.start, b.start);
          default: return 0;
        }
      });
    }

    this.dataSource.sortData = sortConfig;
    this.unacknowledgedDataSource.sortData = sortConfig;
  }

  ngAfterViewInit() {
    this.dataSource.paginator = this.akwPaginator;
    this.dataSource.sort = this.akwSort;

    this.unacknowledgedDataSource.paginator = this.unakwPaginator;
    this.unacknowledgedDataSource.sort = this.unakwSort;
  }

  ngOnDestroy(): void {
    if (this.issueUpdateInterval) {
      window.clearInterval(this.issueUpdateInterval);
    }
  }

  private updateIssues(): void {
    var unacknowledgedIssues : Issue[] = [];
    var acknowledgedIssues : Issue[] = [];

    const now = new Date();
    const time = 2 * 60 * 1000; // 2 min in milliseconds
    this.api.getIssues().subscribe(issues => {
      for (let index = 0; index < issues.length; index++) {
        let minimumReached = false;
        issues[index].alerts?.forEach(a => {
          let start = new Date(a.start);
          if (now.getTime() - start.getTime() > time) {
            minimumReached = true;
          }
        });

        if (minimumReached) { // suppress alerts younger than 2 min
          if(issues[index].acknowledgedTime != undefined){
            acknowledgedIssues.push(issues[index]);
          }else{
            unacknowledgedIssues.push(issues[index]);
          }
        }
      }

      
      this.dataSource.data = acknowledgedIssues;
      this.unacknowledgedDataSource.data = unacknowledgedIssues;
      this.totalIssues = issues.length;
      this.totalAlerts = 0;
      for (let index = 0; index < issues.length; index++) {
        this.totalAlerts += this.getActiveAlerts(issues[index]);
      }
    })
  }

  acknowledgeIssue(issue: Issue): void {
    this.api.acknowledgeIssue(issue).subscribe(issue => {
    }, err => {
      console.log("unable to link issue", err);
    });
    this.updateIssues();
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

  applyFilter() {
    if(this.filterValue === ""){
      this.dataSource.filter = "◬";
      this.unacknowledgedDataSource.filter = "◬";
    }else{
      const filterValue = this.filterValue;
      var filters = filterValue.split(" ", 10);
      for(var word of filters){
        console.log(word);
        this.dataSource.filter = word.trim().toLowerCase();
        this.unacknowledgedDataSource.filter = word.trim().toLowerCase();
      }
    }
  }

  getActiveAlerts(issue: Issue): number{ // counts the active alerts
    var count = 0;
    if (!issue.alerts) {
      return 0;
    }
    for (const [id, alert] of issue?.alerts.entries()) {
      if (alert.end) {
        continue;
      }else{
       count++;
      }
    }
    return count;
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
  styleUrls: ['../dialog.scss'],
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
  styleUrls: ['../dialog.scss'],
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
