import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, BehaviorSubject, throwError } from 'rxjs';
import { finalize, catchError } from 'rxjs/operators';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import jwt_decode, { JwtPayload } from 'jwt-decode';

import { BaseService } from './base.service';
import { DialogService } from './dialog.service';

interface JWTData {
  email: string;
  username: string;
}

interface RegisterResponse {
  requiresVerification: boolean
  alreadyVerified: boolean
}

@Injectable({
  providedIn: 'root',
})
export class AuthenticationService extends BaseService {
  // apis for the authentication service
  requestResetPasswordApiUrl = '/api/reset';
  loginApiUrl = '/api/login';
  logoutApiUrl = '/api/logout';
  registerApiUrl = '/api/register';
  refreshApiUrl = '/api/refresh';

  // the local storage key for tokens
  TOKEN_KEY = 'token';

  // UI can subscribe to this to reflect authentication state
  isAuthenticated$: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(false);

  // crude indicator to the http interceptor that it is to suspend redirects to login...
  suspendLoginRedirects: boolean = false

  constructor(
    private router: Router,
    private dialogService: DialogService,
    private http: HttpClient
  ) {
    super();
    // notify observers that we think we're authenticated
    if (this.token) {
      this.isAuthenticated$.next(true);
    }
  }

  /**
   * clear the stored token
   */
  private clearToken(): void {
    localStorage.removeItem(this.TOKEN_KEY)
    this.isAuthenticated$.next(false)
  }

  /**
   * set the token
   * @param token token string
   */
  public setToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
    this.isAuthenticated$.next(true)
  }

  /**
   *
   * @returns The username we're logged in with.
   */
  public get username(): string {
    if (!this.token) {
      return '';
    }
    try {
      return (jwt_decode(this.token) as JWTData).username;
    } catch (Error) {
      console.log('error decoding token');
      return '';
    }
  }
  /**
   *
   * @returns The email address we're logged in with.
   */
  public get email(): string {
    if (!this.token) {
      return '';
    }
    try {
      return (jwt_decode(this.token) as JWTData).email;
    } catch (Error) {
      console.log('error decoding token');
      return '';
    }
  }

  /**
   * 
   * @returns True if the current token is expired. 
   * @returns False if the current token is not set.
   */
  public isTokenExpired(): boolean {
    if (!this.token) {
      // non existent tokens cannot be expired
      return false;
    }
    const decodedToken: any = jwt_decode(this.token);
    const currentTime = Date.now() / 1000; // convert to seconds
    return decodedToken.exp < currentTime;
  }

  /**
   * 
   * @returns the token's expiration time, or the empty string if the token is unset.
   */
  public getExpirationTime(): string {
    if (!this.token) {
      return '';
    }
    const decodedToken: JwtPayload = jwt_decode(this.token)
    if (!decodedToken || !decodedToken.exp) {
      return '';
    }
    const expDate = new Date(0)
    expDate.setUTCSeconds(decodedToken.exp);

    const expTime = expDate.toLocaleString('en-US', {
      timeZone: 'America/New_York',
      hour12: true,
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
      second: 'numeric'
    });
    return expTime
  }

  /**
   * obtain the token
   */
  public get token(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  /**
   * 
   * @param modalText optional modal text to display upon logout.
   * @param redirectToLogin whether or not to redirect the user to the login page
   * when they log out.
   */
  public logout(modalText?: string, redirectToLogin?: boolean): void {
    this.logoutHttp().subscribe((x) => {
      console.log("Logged out.")
    })
    this.clearToken()
    if (modalText) {
      this.dialogService.displayLogDialog(modalText);
    }
    if (redirectToLogin) {
      console.log("Redirecting to login via logout...")
      this.router.navigateByUrl('/authentication?mode=login');
      return;
    }
    this.router.navigateByUrl('/');
  }

  /**
   * 
   * @returns an observable for refreshing the token.
   */
  public refresh(): Observable<any> {
    return this.refreshHttp()
  }

  /**
   * this uses the reset code sent from the first step of this workflow to 
   * set the new password.
   * @param password new password
   * @param passwordConfirmation confirm new password
   * @param resetCode the reset code for authentication
   * @returns 
   */
  public performPasswordReset(email: string,
    resetCode: string,
    password: string,
    passwordConfirmation: string): void {
    this.suspendLoginRedirects = true
    this.isLoading$.next(true);
    this.performPasswordResetHttp(email, resetCode, password, passwordConfirmation)
      .pipe(
        catchError((error: any) => {
          if (error instanceof HttpErrorResponse) {
            this.error$.next(error.error.errorMessage);
            this.errorCode$.next(0); // send a benign event so observers can close modals
          } else {
            this.error$.next('Unexpected error');
            this.errorCode$.next(0); // send a benign event so observers can close modals
          }
          return throwError(() => new Error("Failed requesting password reset"));
        }),
        finalize(() => { this.isLoading$.next(false), this.suspendLoginRedirects = false })
      )
      .subscribe((x: any) => {
        this.error$.next(''); // send a benign event so observers can close modals
        this.errorCode$.next(0); // send a benign event so observers can close modals
        this.router.navigateByUrl(`/authentication?mode=login&resetSuccessful=true`);
      });
    return
  }

  /**
   * request the backend to begin the password reset workflow.
   * @param email email requesting the password reset.
   */
  public requestPasswordReset(email: string): void {
    this.suspendLoginRedirects = true
    this.isLoading$.next(true);
    this.requestPasswordResetHttp(email)
      .pipe(
        catchError((error: any) => {
          if (error instanceof HttpErrorResponse) {
            this.error$.next(error.error.errorMessage);
            this.errorCode$.next(0); // send a benign event so observers can close modals
          } else {
            this.error$.next('Unexpected error');
            this.errorCode$.next(0); // send a benign event so observers can close modals
          }
          return throwError(() => new Error("Failed requesting password reset"));
        }),
        finalize(() => { this.isLoading$.next(false); this.suspendLoginRedirects = false })
      )
      .subscribe((x: any) => {
        this.error$.next(''); // send a benign event so observers can close modals
        this.errorCode$.next(0); // send a benign event so observers can close modals
        this.router.navigateByUrl(`/authentication?mode=performPasswordReset`);
      });

  }

  /**
   * 
   * @param email the email to register with
   * @param username the username to register with
   * @param password the password to register with
   */
  public register(
    email: string,
    username: string,
    password: string,
  ) {
    this.isLoading$.next(true)
    this.registerHttp(email, username, password)
      .pipe(
        catchError((error: any) => {
          if (error instanceof HttpErrorResponse) {
            this.error$.next(error.error.errorMessage);
            this.errorCode$.next(0); // send a benign event so observers can close modals
          } else {
            this.error$.next('Unexpected error');
            this.errorCode$.next(0); // send a benign event so observers can close modals
          }
          return throwError(() => new Error("Failed registering"));
        }),
        finalize(() => this.isLoading$.next(false))
      )
      .subscribe((x: RegisterResponse) => {
        console.log('registration completed OK: ' + x.requiresVerification);
        this.error$.next(''); // send a benign event so observers can close modals
        this.errorCode$.next(0); // send a benign event so observers can close modals
        this.router.navigateByUrl(`/authentication?mode=login&requiresVerification=${x.requiresVerification}`);
      });
  }

  /**
   * 
   * @param email the email to use for logging in
   * @param password the password to use for logging in
   */
  public login(email: string, password: string, resendCode?: boolean) {
    this.isLoading$.next(true)
    this.loginHttp(email, password, resendCode)
      .pipe(
        catchError((error: any) => {
          if (error instanceof HttpErrorResponse) {
            this.error$.next(error.error.errorMessage);
            this.errorCode$.next(error.error.errorCode); // send a benign event so observers can close modals
          } else {
            this.error$.next('Unexpected error');
            this.errorCode$.next(0); // send a benign event so observers can close modals
          }
          return throwError(() => new Error("Failed logging in"));
        }),
        finalize(() => this.isLoading$.next(false))
      )
      .subscribe((x) => {
        if (x.token == undefined && x.message != undefined) {
          console.log("Got something other than a token...")
          this.error$.next(''); // send a benign event so observers can close modals
          this.errorCode$.next(0); // send a benign event so observers can close modals
          this.router.navigateByUrl(`/authentication?mode=login&codeResent=true`);
          return
        }
        console.log('Setting token to ' + x.token);
        this.setToken(x.token)
        this.error$.next(''); // send a benign event so observers can close modals
        this.errorCode$.next(0); // send a benign event so observers can close modals
        this.router.navigateByUrl('/');
      });
  }

  /**
   * obtain a new access token by sending the browser cookie refresh-token.
   * @returns an observable for obtaining a new access token.
   */
  private refreshHttp(): Observable<any> {
    return this.http.get(this.getUrlBase() + this.refreshApiUrl, this.httpOptions)
  }

  /**
   * register a new account
   * @param email email to register with
   * @param username username to register with
   * @param password password to register with
   * @returns 
   */
  private registerHttp(
    email: string,
    username: string,
    password: string,
  ): Observable<any> {
    const data = {
      email: email,
      username: username,
      password: password,
    };
    return this.http.post(
      this.getUrlBase() + this.registerApiUrl,
      data,
      this.httpOptions
    );
  }

  /**
   * log out.
   * @returns an observable for completing the logout workflow.
   */
  private logoutHttp(): Observable<any> {
    return this.http.post(this.getUrlBase() + this.logoutApiUrl, null, this.httpOptions)
  }

  /**
   * 
   * @param email email to log in with
   * @param password password to log in with
   * @returns 
   */
  private loginHttp(email: string, password: string, resendCode?: boolean): Observable<any> {
    const data = {
      email: email,
      password: password,
    };

    if (resendCode) {

      return this.http.post(
        this.getUrlBase() + this.loginApiUrl + "?resend=true",
        data,
        this.httpOptions
      );
    }
    return this.http.post(
      this.getUrlBase() + this.loginApiUrl,
      data,
      this.httpOptions
    );
  }
  private requestPasswordResetHttp(email: string): Observable<any> {
    const data = {
      email: email,
    };
    return this.http.post(
      this.getUrlBase() + this.requestResetPasswordApiUrl,
      data,
      this.httpOptions
    );
  }

  private performPasswordResetHttp(email: string, resetCode: string, password: string, passwordConfirmation: string): Observable<any> {
    const data = {
      email: email,
      resetCode: resetCode,
      password: password,
      passwordConfirmation: passwordConfirmation,
    };
    return this.http.post(
      this.getUrlBase() + this.requestResetPasswordApiUrl + "?doComplete=true",
      data,
      this.httpOptions
    );
  }
}
