import { Component, Input } from '@angular/core';
import { EventEmitter } from '@angular/core';
import { Output } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

import { AuthenticationService } from 'src/app/services/authentication.service';
import { DialogService, PlantCareDialogData, PlantNotesDialogData } from 'src/app/services/dialog.service';
import { PlantsService } from 'src/app/services/plants.service';
import { Plant } from 'src/app/models/plant.model';

@Component({
  selector: 'app-plant',
  templateUrl: './plant.component.html',
  styleUrls: ['./plant.component.css']
})
export class PlantComponent {
  panelOpenState: boolean = false;

  // passed from the parent component
  @Input() plant: Plant | undefined

  // whether or not we're supposed to take up less screen space per plant
  @Input() isCondensed?: boolean

  // whether or not this plant is overdue for watering
  needsWatering: boolean = false

  // whether or not this plant is overdue for fertilizing
  needsFertilizing: boolean = false

  // whether or not the image for this plant is in progress loading
  isImageLoading: boolean = false

  // filled out by getImage when the image loads - this is the string to render 
  // at the frontend with <img [src]="imageUrl">
  imageUrl: string | null = null

  // when user clicks the "edit" button on this plant
  @Output() editModeEmitter = new EventEmitter<{ plant: Plant, imageUrl: string | null }>()

  // color for text for the 'next care date' - set to red for plants in need of care
  backgroundColorMoist: string = 'orange';
  backgroundColorWater: string = 'black';
  backgroundColorFertilize: string = 'black';

  numComments$: BehaviorSubject<number> = new BehaviorSubject<number>(0)

  constructor(
    private plantsService: PlantsService,
    private dialogService: DialogService,
    public authenticationService: AuthenticationService) {
    this.isCondensed = false // deafult value

  }

  ngOnInit() {
    this.getImage()
    if (new Date(this.getNextFertilizeDate()) < new Date()) {
      if (this.plant) {
        if (this.plant.fertilizingFrequency > 0) {
          this.backgroundColorFertilize = "red"
          this.needsFertilizing = true
        }
      }
    }
    if (new Date(this.getNextWaterDate()) < new Date() && this.plant?.lastMoistDate == '') {
      if (this.plant) {
        this.backgroundColorWater = "red"
        this.needsWatering = true
      }
    }
    if (new Date(this.getNextMoistCheckDate()) < new Date()) {
      if (this.plant) {
        this.backgroundColorMoist = "red"
      }
    }
    if (this.plant && this.authenticationService.isAuthenticated$.value && this.plant.username == this.authenticationService.username) {
      let unviewedComments = this.plant?.comments.filter(x =>
        !x.viewed
      ).length
      console.log("plantId=" + this.plant.ID + " has " + unviewedComments + " unviewed comments.")
      this.numComments$.next(unviewedComments)
    }
  }

  /**
   * Handle user editing the plant. Emits an event for the plants component HTML
   * to listen for.
   */
  public editPlant(): void {
    if (!this.plant) {
      return;
    }
    this.editModeEmitter.emit({ plant: this.plant, imageUrl: this.imageUrl })
  }

  /**
   * opens the notes modal
   */
  public editPlantNotes(): void {
    if (!this.plant) {
      return;
    }
    console.log("Plant notes are: " + this.plant.notes)
    var dialogRef = this.dialogService.displayNotesDialog(this.plant.notes)
    dialogRef.afterClosed().subscribe((result?: PlantNotesDialogData) => {
      if (!this.plant) {
        return;
      }
      if (!result || result.cancelled) {
        return;
      }
      if (result.notes.length > 1024) {
        this.dialogService.displayErrorDialog("Max note size (" + 1024 + ") exceeded.")
        return
      }
      console.log("Notes saved: " + result.notes)
      this.plant.notes = result.notes
      this.plantsService.updatePlant(this.plant, null)
    })
  }

  /**
   * Handle user hitting "Edit" from the dropdown menu to update plant care
   * status.
   */
  public openCareDialog(): void {
    if (!this.plant) {
      return
    }
    var dialogRef = this.dialogService.displayPlantCareDialog(this.plant.name)
    dialogRef.afterClosed().subscribe((result: PlantCareDialogData) => {
      if (result && result.cancelled) {
        console.log("Edit cancelled.");
        return
      }
      console.log("plant care result: " + JSON.stringify(result))
      if (result.moist && (result.water || result.fertilize)) {
        this.dialogService.displayErrorDialog("Only choose 'moist' when not choosing other care actions.")
        return;
      }
      if (result) {
        if (!this.plant) {
          return;
        }
        if (result.water) {
          this.plant.lastWaterDate = Plant.formatDate(new Date())
          console.log("Setting lastWaterDate to " + this.plant.lastWaterDate);
          this.plant.lastMoistDate = '' // unset
        }
        if (result.fertilize || result.skipFertilize) {
          this.plant.lastFertilizeDate = Plant.formatDate(new Date())
          console.log("Setting lastFertilizeDate to " + this.plant.lastFertilizeDate);
          if (result.skipFertilize) {
            console.log("Fertilize was skipped")
            this.plant.skippedLastFertilize = true
          } else {
            this.plant.skippedLastFertilize = false
          }
        }
        if (result.moist) {
          console.log("Plant is moist - only updating that attribute")
          this.plant.lastMoistDate = Plant.formatDate(new Date())
          // this.plantsService.markMoist(this.plant)
          this.plantsService.updatePlant(this.plant, null)
          return;
        }

        // not updating the image for this plant
        this.plantsService.updatePlant(this.plant, null)
      } else {
        console.log("Dialog declined.")
      }
    })

  }

  /**
   * format the last water date to a string.
   * @returns formatted last water date
   */
  public transformLastPlantCareDate(date: string): string {
    let myDate = new Date(date);
    return Plant.formatDate(myDate)
  }

  /**
   * 
   * @returns logs sorted by their date
   */
  public getSortedLogs(): any {
    var sortedLogs = this.plant?.logs.slice().sort((a, b) => new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime());
    return sortedLogs;
  }

  /**
   * 
   * @param dbDate the date string from gorm, like from gorm.CreatedAt.
   * Example dbDate: 2023-05-24T05:31:40.232118Z
   * @returns mm/dd/yyyy h:m:s AM/PM formatted date string.
   * Example return value: 5/24/2023 5:31:40 AM
   */
  public transformGormDatabaseDate(dbDate: string): string {
    const date = new Date(dbDate);
    let hours = date.getHours();
    const ampm = hours >= 12 ? 'PM' : 'AM';
    hours = hours % 12;
    hours = hours ? hours : 12; // the hour '0' should be '12'
    const minutes = date.getMinutes();
    const strMinutes = minutes < 10 ? '0' + minutes : minutes;
    const formattedDate = `${date.getMonth() + 1}/${date.getDate()}/${date.getFullYear()} ${hours}:${strMinutes}:${date.getSeconds()} ${ampm}`;
    return formattedDate;
  }

  /**
   * Obtain a formatted next-action-date from the difference between this plant's last action date 
   * and action frequency.
   * @param {string} lastActionDate - The last date the action was performed.
   * @param {number} frequency - The frequency of the action in days.
   * @returns {string} The date to next perform the action on this plant.
   */
  public getNextActionDate(lastActionDateStr: string, frequency: number): string {
    if (!this.plant) {
      return "N/A"
    }
    var nextActionDate = new Date()
    var lastActionDate = new Date(lastActionDateStr)
    nextActionDate.setFullYear(lastActionDate.getFullYear());
    nextActionDate.setMonth(lastActionDate.getMonth());
    var frequencyInMs = frequency * 24 * 60 * 60 * 1000;
    nextActionDate.setTime(lastActionDate.getTime() + frequencyInMs);
    return Plant.formatDate(nextActionDate)
  }

  /**
   * Obtain a formatted next-fertilize-date from the difference between this plant's last fertilize date 
   * and fertilize frequency.
   * @returns {string} The date to next fertilize this plant.
   */
  public getNextFertilizeDate(): string {
    return this.getNextActionDate(this.plant!.lastFertilizeDate, this.plant!.fertilizingFrequency);
  }

  /**
   * Obtain a formatted next-water-date from the difference between this plant's last water date 
   * and water frequency.
   * @returns {string} The date to next water this plant.
   */
  public getNextWaterDate(): string {
    return this.getNextActionDate(this.plant!.lastWaterDate, this.plant!.wateringFrequency);
  }

  /**
   * obtain a formatted next-moist-check-date from the difference between this plants last moist date 
   * and one day.
   * @returns the date to next check moist soil for this plant.
   */
  public getNextMoistCheckDate(): string {
    if (!this.plant) {
      return "N/A"
    }
    var nextMoistDate = new Date()
    var lastMoistDate = new Date(this.plant.lastMoistDate)
    nextMoistDate.setFullYear(lastMoistDate.getFullYear());
    nextMoistDate.setMonth(lastMoistDate.getMonth());
    var frequencyInMs = 1 * 24 * 60 * 60 * 1000;
    nextMoistDate.setTime(lastMoistDate.getTime() + frequencyInMs);
    return Plant.formatDate(nextMoistDate)
  }

  /**
   * get the image for this plant based on its imageId
   */
  private getImage(): void {
    if (!this.plant || this.plant.imageId == 0) {
      return
    }
    console.log("Getting image for imageId=" + this.plant.imageId)
    this.isImageLoading = true;
    this.plantsService.getPlantImage(this.plant.imageId)
      .subscribe(blob => {
        const reader = new FileReader();
        reader.onloadend = () => {
          this.imageUrl = reader.result as string;
        };
        reader.readAsDataURL(blob);
        this.isImageLoading = false;
      });
  }

  /**
   * delete this plant. Handle the 'delete' button. 
   */
  public deletePlant() {
    if (!this.plant) {
      return;
    }
    var dialogRef = this.dialogService.displayConfirmationDialog("Are you sure you want to delete plant: " + this.plant.name + "?")
    if (this.plant == null) {
      console.log("Unexpected plant is NULL");
      return;
    }
    dialogRef.afterClosed().subscribe((result: boolean) => {
      if (result) {
        if (!this.plant) {
          return;
        }
        this.plantsService.deletePlant(this.plant.ID);
      } else {
        console.log("Dialog declined.")
      }
    })

  }

}
