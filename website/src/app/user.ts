import {CookieService} from 'ngx-cookie-service';
import {JwtHelperService} from '@auth0/angular-jwt';

export class User {
    username: string;

    constructor(private cookie: CookieService) {

        this.username = "";

        const decoder = new JwtHelperService();
        var token = decoder.decodeToken(this.cookie.get("smee"));
        if (token != null) {
            this.username = token.user;
        }
    }
}


