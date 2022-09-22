import {Component, OnInit} from '@angular/core';
import {ApiService, RoomOverview} from "../api.service";

@Component({
  selector: 'app-rooms',
  templateUrl: './rooms.component.html',
  styleUrls: ['./rooms.component.scss']
})
export class RoomsComponent implements OnInit {
  filter: string = "";
  onlyMaintenance: boolean = false;
  rooms: RoomOverview[] = [];

  constructor(private api: ApiService) {}

  ngOnInit(): void {
    this.update();
  }

  update(): void {
    this.api.getRooms().subscribe(rooms => {
      this.rooms = rooms;
    })
  }

  get filtered(): RoomOverview[] {
    return this.rooms.filter(room => {
      if (this.onlyMaintenance && !room.inMaintenance) {
        return false;
      }

      if (room.id.toLowerCase().includes(this.filter.toLowerCase())) {
        return true;
      }

      return false;
    });
  }
}
