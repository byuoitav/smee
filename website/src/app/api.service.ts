import {HttpClient, HttpErrorResponse, HttpParams} from "@angular/common/http";
import { stringify } from "@angular/compiler/src/util";
import {Injectable} from '@angular/core';
import { MatPaginator } from "@angular/material/paginator";
import {Observable, of, throwError} from "rxjs";
import {tap, map, catchError} from "rxjs/operators";

export interface Alert {
  id: string;
  issueID: string;
  device: Device;
  type: string;
  start: Date;
  end: Date;
  link: string | undefined;
}

export interface Room {
  id: string;
  name: string;
}

export interface Device {
  id: string;
  name: string;
  room: Room;
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
  room: Room;
  start: Date | undefined;
  end: Date | undefined;
  alerts: Map<string, Alert> | undefined;
  incidents: Map<string, Incident> | undefined;
  events: IssueEvent[] | undefined;
  maintenanceStart: Date | undefined;
  maintenanceEnd: Date | undefined;
  isOnMaintenance : boolean; 
}

export interface MaintenanceInfo {
  roomID: string;
  start: Date | undefined;
  end: Date | undefined;
  note: string | undefined;
}

export interface RoomOverview {
  id: string;
  name: string;
  inMaintenance: boolean;
}

export interface IssueType{
  idAlertType : string
  kbArticle : string
}

export interface IssueTypeMap{
  IssueType: Map<string, IssueType>;
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  constructor(private http: HttpClient) {}

  getRooms(): Observable<RoomOverview[]> {
    return this.http.get<RoomOverview[]>("/api/v1/rooms").pipe(
      tap(data => console.log("got rooms", data)),
      catchError(this.handleError<RoomOverview[]>("getRooms", [])),
    )
  }

  getIssueType(): Observable<IssueTypeMap> {
    return this.http.get<IssueTypeMap>("api/v1/issuetype").pipe(
      tap(data => console.log("got issue type", data)),
      catchError(this.handleError<IssueTypeMap>("getIssueType", )),
      map((issuetype : IssueTypeMap) => {
        const issTypeMap = issuetype.IssueType
        issuetype.IssueType = new Map(Object.entries(issTypeMap));
        return issuetype
      }),
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
          issues[i].isOnMaintenance = this.inMaintenance(issues[i]); //assings value to isOnMaintenance
        }
        return issues;
      }),
    )
  }

  inMaintenance(issue : Issue):  boolean{ // funciton that dtermines if an issue is on Maintenance or not
    if (!issue.maintenanceStart || !issue.maintenanceEnd){
      return false;
    }else{
    const now = new Date();
      if (now < issue.maintenanceStart){
      return false;
      } else if (now > issue.maintenanceEnd){
        return false;
      }
    }
    return true;
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


