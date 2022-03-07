import {Component, OnInit} from '@angular/core';
import {CookieService} from 'ngx-cookie-service';
import {JwtHelperService} from '@auth0/angular-jwt';

export interface User {
  username: string;
}


@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit{
  user: User;

  constructor(private cookieService: CookieService) {
    this.user = {username: "netID"};
  }

  ngOnInit(): void {
    const decoder = new JwtHelperService();
    var token = decoder.decodeToken(this.cookieService.get("smee"))
    if (token != null) {
      this.user.username = token.user
    }
  }
}
