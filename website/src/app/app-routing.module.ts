import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import { CommandsComponent } from './commands/commands.component';
import {DashboardComponent} from "./dashboard/dashboard.component";
import { LogoutComponent } from './logout/logout.component';
import {RoomComponent} from "./room/room.component";
import {RoomsComponent} from "./rooms/rooms.component";

const routes: Routes = [
  {
    path: '',
    redirectTo: '/dashboard',
    pathMatch: 'full'
  },
  {
    path: 'dashboard',
    component: DashboardComponent
  },
  {
    path: 'rooms',
    component: RoomsComponent
  },
  {
    path: 'rooms/:roomID',
    component: RoomComponent
  },
  {
    path: 'commands',
    component: CommandsComponent
  },
  {
    path: 'logout',
    component: LogoutComponent
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule {}
