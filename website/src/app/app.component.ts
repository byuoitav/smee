import {Component, OnInit} from '@angular/core';
import {CookieService} from 'ngx-cookie-service';
import { User } from './user';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit{
  user: User;

  constructor(private cS: CookieService) {
    this.user = new User(cS);
  }

  ngOnInit(): void {}
}
