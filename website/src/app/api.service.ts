import {HttpClient} from "@angular/common/http";
import {Injectable} from '@angular/core';
import {Observable, of} from "rxjs";
import {tap, map, catchError} from "rxjs/operators";

export interface Alert {
  id: string;
  issueID: string;
  room: string;
  device: string; type: string;
  start: Date;
  end: Date;
}

export interface Incident {
  id: string;
  name: string;
}

export interface IssueEvent {
  timestamp: Date;
  type: string;
  data: string;
}

export interface Issue {
  id: string;
  room: string;
  start: Date;
  end: Date;
  alerts: Map<string, Alert>;
  incidents: Map<string, Incident>;
  events: Map<string, IssueEvent>;
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  constructor(private http: HttpClient) {}

  getIssues(): Observable<Issue[]> {
    return this.http.get<Issue[]>("/api/v1/issues").pipe(
      tap(data => console.log("got issues", data)),
      catchError(this.handleError<Issue[]>("getIssues", [])),
      map((issues: Issue[]) => {
        for (let i in issues) {
          issues[i].alerts = new Map<string, Alert>(Object.entries(issues[i].alerts));
          issues[i].incidents = new Map<string, Incident>(Object.entries(issues[i].incidents));
          issues[i].events = new Map<string, IssueEvent>(Object.entries(issues[i].events));
        }

        return issues;
      }),
    )
  }

  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.error(error);
      return of(result as T);
    };
  }
}
