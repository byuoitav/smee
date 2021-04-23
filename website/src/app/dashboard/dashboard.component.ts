import {Component, OnInit} from '@angular/core';
import {MatTableDataSource} from "@angular/material/table";
import {ApiService, Issue, Alert, Incident, IssueEvent} from "../api.service";

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  displayedColumns: string[] = ["room", "alertCount", "started", "incidents"];
  dataSource: MatTableDataSource<Issue> = new MatTableDataSource(undefined);

  constructor(private api: ApiService) {}

  ngOnInit(): void {
    this.updateIssues();
    setInterval(() => {
      this.updateIssues();
    }, 10000)
  }

  private updateIssues() {
    this.api.getIssues().subscribe(issues => {
      this.dataSource = new MatTableDataSource(issues);
    })
  }
}
