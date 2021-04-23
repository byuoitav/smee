import {Component, OnInit} from '@angular/core';
import {MatTableDataSource} from "@angular/material/table";

export interface Alert {
}

export interface Incident {
}

export interface IssueEvent {
}

export interface Issue {
  id: string;
  room: string;
  start: Date;
  end: Date;
  alerts: Map<string, Alert>;
  incidents: Map<string, Incident>;
  events: Map<string, IssueEvent>;
}

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  displayedColumns: string[] = ["room"];
  dataSource: MatTableDataSource<Issue>

  constructor() {
    const issues: Issue[] = [
      {
        id: "one",
        room: "ITB-1101",
        start: new Date(),
        end: new Date(),
        alerts: new Map<string, Alert>(),
        incidents: new Map<string, Incident>(),
        events: new Map<string, IssueEvent>(),
      },
    ];
    this.dataSource = new MatTableDataSource(issues);
  }

  ngOnInit(): void {}
}
