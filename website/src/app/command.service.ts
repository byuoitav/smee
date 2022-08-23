import { Injectable } from '@angular/core';
import { HttpClient } from "@angular/common/http";
import { Observable, of } from "rxjs";
import { tap, catchError, map } from "rxjs/operators";
import { ApiService } from './api.service';
import { User } from './user';
import { CookieService } from 'ngx-cookie-service';

@Injectable({
  providedIn: 'root'
})
export class CommandService {
  user: User;

  constructor(private http: HttpClient,
    private api: ApiService,
    private cookie: CookieService) {
      this.user = new User(cookie);
    }

  float(input: string): Observable<any> {
    return this.http.put<any>("/api/v1/commands/float/" + input, undefined);
  }

  swab(input: string): Observable<any> {
    return this.http.put<string[]>("/api/v1/commands/swab/" + input, undefined);
  }

  sink(input: string): Observable<any> {
    return this.http.put<any>("/api/v1/commands/sink/" + input, undefined);
  }

  fixTime(input: string): Observable<any> {
    return this.http.put<any>("/api/v1/commands/fixTime/" + input, undefined);
  }

  rmDevice(input: string): Observable<any> {
    return this.http.put<any>("/api/v1/commands/removeDevice/" + input, undefined);
  }

  closeIssue(input: string): Observable<any> {
    //find the issue id from the room id
    let issueID = "";
    this.api.getIssues().subscribe(issues => {
      for (let i = 0; i < issues.length; i++) {
        if (issues[i].room.name == input) {
          issueID = issues[i].id;
        }
      }
    });

    if (issueID != "") {
      this.api.closeIssue(issueID); 
    }

    return this.http.put<any>("/api/v1/commands/closeIssue/" + input, undefined);
  }

  dupDatabase(src: string, dest: string): Observable<any> {
    return this.http.put<any>("/api/v1/commands/duplicateDatabase/" + src + "/" + dest, undefined);
  }

  screenshot(input: string): Observable<any> {
    return this.http.put<any>("/api/v1/commands/screenshot/" + input, undefined);
    // get screenshot and display... somehow
  }

  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.error(`error doing ${operation}`, error);
      return of(result as T);
    };
  }

}
