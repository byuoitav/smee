import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class CommandService {

  constructor() { }

  float(input: string): boolean {
    // call float enpoint
    return false
  }

  swab(input: string): boolean {
    // call swab enpoint
    return false
  }

  sink(input: string): boolean {
    // call sink enpoint
    return false
  }

  fixTime(input: string): boolean {
    // call fixTime enpoint
    return false
  }

  rmDevice(input: string): boolean {
    // call rmDevice enpoint
    return false
  }

  closeIssue(input: string): boolean {
    // call closeIssue enpoint
    return false
  }

  dupeDatabase(src: string, dest: string): boolean {
    // call dupDatabase enpoint
    return false
  }

  screenshot(input: string) {
    // call screenshot endpoint
    // get screenshot and display... somehow
  }

}
