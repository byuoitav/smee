import {Component, OnInit} from '@angular/core';
import {ApiService, Room} from "../api.service";

@Component({
  selector: 'app-rooms',
  templateUrl: './rooms.component.html',
  styleUrls: ['./rooms.component.scss']
})
export class RoomsComponent implements OnInit {
  filter: string = "";
  onlyMaintenance: boolean = false;
  rooms: Room[] = [];

  constructor(private api: ApiService) {}

  ngOnInit(): void {
    this.update();
  }

  update(): void {
    this.api.getRooms().subscribe(rooms => {
      this.rooms = rooms;
    })
  }

  get filtered(): Room[] {
    return this.rooms.filter(room => {
      if (this.onlyMaintenance && !room.inMaintenance) {
        return false;
      }

      if (room.name.toLowerCase().includes(this.filter.toLowerCase())) {
        return true;
      }

      if (room.id.toLowerCase().includes(this.filter.toLowerCase())) {
        return true;
      }

      return false;
    });
  }
}
