import {Component} from '@angular/core';

export interface User {
  username: string;
}

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  user: User;

  constructor() {
    this.user = {username: "netID"};
  }
}
