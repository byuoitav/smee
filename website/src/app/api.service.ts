import {HttpClient, HttpErrorResponse, HttpParams} from "@angular/common/http";
import {Injectable} from '@angular/core';
import {Observable, of, throwError} from "rxjs";
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
  alerts: Map<string, Alert> | undefined;
  incidents: Map<string, Incident> | undefined;
  events: IssueEvent[] | undefined;
  maintenanceStart: Date | undefined;
  maintenanceEnd: Date | undefined;
}

export interface MaintenanceInfo {
  roomID: string;
  start: Date | undefined;
  end: Date | undefined;
}

export interface Room {
  id: string;
  name: string;
  inMaintenance: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  constructor(private http: HttpClient) {}

  getRooms(): Observable<Room[]> {
    return this.http.get<Room[]>("/api/v1/rooms").pipe(
      tap(data => console.log("got rooms", data)),
      catchError(this.handleError<Room[]>("getRooms", [])),
    )
  }

  getIssues(): Observable<Issue[]> {
    return this.http.get<Issue[]>("/api/v1/issues").pipe(
      // tap(data => console.log("got issues", data)),
      catchError(this.handleError<Issue[]>("getIssues", [])),
      map((issues: Issue[]) => {
        for (let i in issues) {
          const issue = issues[i]; // apparently the ts compiler needs this to be happy

          if (issue?.alerts) {
            issues[i].alerts = new Map(Object.entries(issue.alerts));
          }

          if (issue?.incidents) {
            issues[i].incidents = new Map(Object.entries(issue.incidents));
          }
        }

        return issues;
      }),
    )
  }

  getIssue(roomID: string): Observable<Issue> {
    return this.http.get<Issue>("/api/v1/issues", {
      params: new HttpParams().set("roomID", roomID)
    }).pipe(
      // tap(data => console.log("got issue", data)),
      catchError(this.handleError<Issue>("getIssue", undefined)),
      map((issue: Issue) => {
        if (issue?.alerts) {
          issue.alerts = new Map(Object.entries(issue.alerts));
        }

        if (issue?.incidents) {
          issue.incidents = new Map(Object.entries(issue.incidents));
        }

        return issue;
      }),
    )
  }

  linkIssueToIncident(issueID: string, incName: string): Observable<Issue> {
    return this.http.put<Issue>(`/api/v1/issues/${issueID}/linkIncident`, undefined, {
      params: new HttpParams().set("incName", incName)
    }).pipe(
      tap(data => console.log("linkIssueToIncident response", data)),
      map((issue: Issue) => {
        if (issue?.alerts) {
          issue.alerts = new Map(Object.entries(issue.alerts));
        }

        if (issue?.incidents) {
          issue.incidents = new Map(Object.entries(issue.incidents));
        }

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
        if (issue?.alerts) {
          issue.alerts = new Map(Object.entries(issue.alerts));
        }

        if (issue?.incidents) {
          issue.incidents = new Map(Object.entries(issue.incidents));
        }

        return issue;
      }),
    );
  }

  getMaintenanceInfo(roomID: string): Observable<MaintenanceInfo> {
    return this.http.get<MaintenanceInfo>(`/api/v1/maintenance/${roomID}`).pipe(
      // tap(data => console.log("got maintenance info", data)),
      catchError(this.handleError<MaintenanceInfo>("getMaintenanceInfo", undefined)),
      map((info: MaintenanceInfo) => {
        if (info?.start) {
          info.start = new Date(info.start);
        }

        if (info?.end) {
          info.end = new Date(info.end);
        }

        return info;
      })
    )
  }

  setMaintenanceInfo(info: MaintenanceInfo): Observable<MaintenanceInfo> {
    return this.http.put<MaintenanceInfo>(`/api/v1/maintenance/${info.roomID}`, info).pipe(
      tap(data => console.log("set maintenance info", data)),
      catchError(this.handleError<MaintenanceInfo>("setMaintenanceInfo", undefined)),
      map((info: MaintenanceInfo) => {
        if (info?.start) {
          info.start = new Date(info.start);
        }

        if (info?.end) {
          info.end = new Date(info.end);
        }

        return info;
      })
    )
  }

  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.error(`error doing ${operation}`, error);
      return of(result as T);
    };
  }
}
