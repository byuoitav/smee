import {HttpClient, HttpParams} from "@angular/common/http";
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
  data: any;
}

export interface Issue {
  id: string;
  room: string | undefined;
  start: Date | undefined;
  end: Date | undefined;
  alerts: Map<string, Alert>;
  incidents: Map<string, Incident>;
  events: IssueEvent[];
  maintenanceStart: Date | undefined;
  maintenanceEnd: Date | undefined;
}

const emptyIssue = (): Issue => ({
  id: '',
  room: undefined,
  start: undefined,
  end: undefined,
  alerts: new Map(),
  incidents: new Map(),
  events: [],
  maintenanceStart: undefined,
  maintenanceEnd: undefined,
})

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
          issues[i].alerts = new Map(Object.entries(issues[i].alerts));
          issues[i].incidents = new Map(Object.entries(issues[i].incidents));
        }

        return issues;
      }),
    )
  }

  getIssue(roomID: string): Observable<Issue> {
    return this.http.get<Issue>("/api/v1/issues", {
      params: new HttpParams().set("roomID", roomID)
    }).pipe(
      tap(data => console.log("got issue", data)),
      map((issues: Issue) => {
        issues.alerts = new Map(Object.entries(issues.alerts));
        issues.incidents = new Map(Object.entries(issues.incidents));

        return issues;
      }),
    )
  }

  linkIssueToIncident(issueID: string, incName: string): Observable<Issue> {
    return this.http.put<Issue>(`/api/v1/issues/${issueID}/linkIncident`, undefined, {
      params: new HttpParams().set("incName", incName)
    }).pipe(
      tap(data => console.log("linkIssueToIncident response", data)),
      map((issue: Issue) => {
        issue.alerts = new Map(Object.entries(issue.alerts));
        issue.incidents = new Map(Object.entries(issue.incidents));
        return issue;
      }),
    );
  }

  createIncidentFromIssue(issueID: string, shortDescription: string): Observable<Issue> {
    return this.http.put<Issue>(`/api/v1/issues/${issueID}/createIncident`, undefined, {
      params: new HttpParams().set('shortDescription', shortDescription)
    }).pipe(
      tap(data => console.log("createIncidentFromIssue response", data)),
      map((issue: Issue) => {
        issue.alerts = new Map(Object.entries(issue.alerts));
        issue.incidents = new Map(Object.entries(issue.incidents));
        return issue;
      }),
    );
  }

  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.error(`error doing ${operation}`, error);
      return of(result as T);
    };
  }
}
