///<reference path="../../headers/common.d.ts" />

export class DashLoader {
  dashboard: any;
  defered: any;

  constructor() {
    this.dashboard = ["$q", "$route", "$rootScope", ($q, $route, $rootScope) => {
      return $q(function(resolve, reject) {
        setTimeout(function() {
          resolve(123);
        }, 1000);
      });
    }];

  }
}
