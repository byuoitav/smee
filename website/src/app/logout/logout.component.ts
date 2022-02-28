import { Component, OnInit } from '@angular/core';
import { CookieService } from 'ngx-cookie-service';

@Component({
  selector: 'app-logout',
  template: ``
})
export class LogoutComponent implements OnInit {

  constructor(private cookieService: CookieService) { }

  ngOnInit(): void {
    this.cookieService.delete('smee');
    window.location.href = "https://api.byu.edu/logout";
  }

}
