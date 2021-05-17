import {Component, OnDestroy, OnInit} from '@angular/core';
import {MatTableDataSource} from "@angular/material/table";
import {Alert, ApiService, Issue} from "../api.service";

@Component({
  selector: 'app-room',
  templateUrl: './room.component.html',
  styleUrls: ['./room.component.scss']
})
export class RoomComponent implements OnInit, OnDestroy {
  roomID: string = "";
  room: string = "";
  issue: Issue | undefined;
  alertsDataSource: MatTableDataSource<Alert> = new MatTableDataSource(undefined);
  alertColumns: string[] = ["device", "type", "started", "ended"];
  updateInterval: number | undefined;

  constructor(private api: ApiService) {}

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
      this.issue.events.push(this.issue.events[0]);
      this.alertsDataSource = new MatTableDataSource([...issue.alerts.values()]);
    });
  }
}
