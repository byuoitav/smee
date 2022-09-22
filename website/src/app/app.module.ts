import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {FormsModule} from '@angular/forms';
import {AppRoutingModule} from './app-routing.module';
import {AppComponent} from './app.component';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {CookieService} from 'ngx-cookie-service';
// angular material
import {MatToolbarModule} from '@angular/material/toolbar';
import {MatIconModule} from '@angular/material/icon';
import {MatSidenavModule} from '@angular/material/sidenav';
import {MatButtonModule} from '@angular/material/button';
import {MatTableModule} from '@angular/material/table';
import {MatFormFieldModule} from '@angular/material/form-field';
import {MatPaginatorModule} from '@angular/material/paginator';
import {MatGridListModule} from '@angular/material/grid-list';
import {MatInputModule} from '@angular/material/input';
import {MatDialogModule} from '@angular/material/dialog';
import {MatCardModule} from '@angular/material/card';
import {MatChipsModule} from '@angular/material/chips';
import {MatSortModule} from "@angular/material/sort";
import {MatCheckboxModule} from '@angular/material/checkbox'; 
import {MatMenuModule} from '@angular/material/menu';
import {MatDividerModule} from '@angular/material/divider';
import {MatProgressBarModule} from '@angular/material/progress-bar';
import {MatTooltipModule} from '@angular/material/tooltip';
import {MatSnackBarModule} from '@angular/material/snack-bar';
//app elements
import {MaintenanceDialog, RoomComponent, CloseIssueDialog, ErrorPopup, StatusDialog} from './room/room.component';
import {DashboardComponent, DashboardCreateDialog, DashboardLinkDialog} from './dashboard/dashboard.component';
import {RoomsComponent} from './rooms/rooms.component';
import {DateAgoPipe} from "./date-ago.pipe";
import { LogoutComponent } from './logout/logout.component';
import { CommandsComponent } from './commands/commands.component';

@NgModule({
  declarations: [
    AppComponent,
    DashboardComponent,
    DashboardCreateDialog,
    DashboardLinkDialog,
    RoomComponent,
    RoomsComponent,
    MaintenanceDialog,
    DateAgoPipe,
    CloseIssueDialog,
    ErrorPopup,
    StatusDialog,
    LogoutComponent,
    CommandsComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    HttpClientModule,
    FormsModule,
    MatToolbarModule,
    MatIconModule,
    MatSidenavModule,
    MatButtonModule,
    MatCheckboxModule,
    MatTableModule,
    MatFormFieldModule,
    MatPaginatorModule,
    MatDialogModule,
    MatInputModule,
    MatCardModule,
    MatGridListModule,
    MatChipsModule,
    MatSortModule,
    MatTableModule,
    MatMenuModule,
    MatDividerModule,
    MatProgressBarModule,
    MatTooltipModule,
    MatSnackBarModule,
  ],
  providers: [CookieService],
  bootstrap: [AppComponent]
})
export class AppModule {}
