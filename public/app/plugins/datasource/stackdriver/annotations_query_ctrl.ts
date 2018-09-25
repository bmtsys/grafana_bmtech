import { FilterSegments } from './filter_segments';
import appEvents from 'app/core/app_events';
import _ from 'lodash';

export class StackdriverAnnotationsQueryCtrl {
  static templateUrl = 'partials/annotations.editor.html';
  annotation: any;
  datasource: any;
  filterSegments: any;
  metricDescriptors: any[];
  metrics: any[];
  loadLabelsPromise: Promise<any>;
  metricLabels: { [key: string]: string[] };
  resourceLabels: { [key: string]: string[] };

  defaultMetricResourcesValue = 'all';
  defaultDropdownValue = 'select metric';

  defaults = {
    project: {
      id: 'default',
      name: 'loading project...',
    },
    metricType: this.defaultDropdownValue,
    metricService: this.defaultMetricResourcesValue,
    metric: '',
    filters: [],
    metricKind: '',
    valueType: '',
  };

  /** @ngInject */
  constructor(private uiSegmentSrv) {
    this.annotation.target = this.annotation.target || {};
    _.defaultsDeep(this.annotation.target, this.defaults);
    this.getCurrentProject()
      .then(this.getMetricTypes.bind(this))
      .then(this.getLabels.bind(this));
    this.initSegments();
  }

  initSegments() {
    this.filterSegments = new FilterSegments(
      this.uiSegmentSrv,
      this.annotation.target,
      this.getFilterKeys.bind(this),
      this.getFilterValues.bind(this)
    );
    this.filterSegments.buildSegmentModel();
  }

  async getCurrentProject() {
    try {
      this.annotation.target.project = await this.datasource.getDefaultProject();
    } catch (error) {
      let message = 'Projects cannot be fetched: ';
      message += error.statusText ? error.statusText + ': ' : '';
      if (error && error.data && error.data.error && error.data.error.message) {
        if (error.data.error.code === 403) {
          message += `
            A list of projects could not be fetched from the Google Cloud Resource Manager API.
            You might need to enable it first:
            https://console.developers.google.com/apis/library/cloudresourcemanager.googleapis.com`;
        } else {
          message += error.data.error.code + '. ' + error.data.error.message;
        }
      } else {
        message += 'Cannot connect to Stackdriver API';
      }
      appEvents.emit('ds-request-error', message);
    }
  }

  async getMetricTypes() {
    if (this.annotation.target.project.id !== 'default') {
      const metricTypes = await this.datasource.getMetricTypes(this.annotation.target.project.id);
      this.metricDescriptors = metricTypes;
      if (this.annotation.target.metricType === this.defaultDropdownValue && metricTypes.length > 0) {
        this.annotation.target.metricType = metricTypes[0].id;
      }

      return metricTypes.map(mt => ({ value: mt.type, text: mt.type }));
    } else {
      return [];
    }
  }

  getMetricServices() {
    const defaultValue = { value: this.defaultMetricResourcesValue, text: this.defaultMetricResourcesValue };
    const resources = (this.metricDescriptors || []).map(m => {
      const [resource] = m.type.split('/');
      const [service] = resource.split('.');
      return {
        value: resource,
        text: service,
      };
    });
    return resources.length > 0 ? [defaultValue, ..._.uniqBy(resources, 'value')] : [];
  }

  getMetrics() {
    const metrics = this.metricDescriptors.map(m => {
      const [resource] = m.type.split('/');
      const [service] = resource.split('.');
      return {
        resource,
        value: m.type,
        service,
        text: m.displayName,
        title: m.description,
      };
    });
    if (this.annotation.target.metricService === this.defaultMetricResourcesValue) {
      return metrics.map(m => ({ ...m, text: `${m.service} - ${m.text}` }));
    } else {
      return metrics.filter(m => m.resource === this.annotation.target.metricService);
    }
  }

  async getLabels() {
    this.loadLabelsPromise = new Promise(async resolve => {
      try {
        const data = await this.datasource.getLabels(this.annotation.target.metricType, 'annotationQuery');

        this.metricLabels = data.results['annotationQuery'].meta.metricLabels;
        this.resourceLabels = data.results['annotationQuery'].meta.resourceLabels;

        this.annotation.target.valueType = data.results['annotationQuery'].meta.valueType;
        this.annotation.target.metricKind = data.results['annotationQuery'].meta.metricKind;
        resolve();
      } catch (error) {
        console.log(error.data.message);
        appEvents.emit('alert-error', [
          'Error',
          'Error loading metric labels for ' + this.annotation.target.metricType,
        ]);
        resolve();
      }
    });
  }

  onResourceTypeChange(resource) {
    this.metrics = this.getMetrics();
    if (!this.metrics.find(m => m.value === this.annotation.target.metricType)) {
      this.annotation.target.metricType = this.defaultDropdownValue;
    }
  }

  async onMetricTypeChange() {}

  async getFilterKeys() {
    await this.loadLabelsPromise;

    const metricLabels = Object.keys(this.metricLabels || {}).map(l => {
      return this.uiSegmentSrv.newSegment({
        value: `metric.label.${l}`,
        expandable: false,
      });
    });

    const resourceLabels = Object.keys(this.resourceLabels || {}).map(l => {
      return this.uiSegmentSrv.newSegment({
        value: `resource.label.${l}`,
        expandable: false,
      });
    });

    // const noValueOrPlusButton = !segment || segment.type === 'plus-button';
    // if (noValueOrPlusButton && metricLabels.length === 0 && resourceLabels.length === 0) {
    //   return Promise.resolve([]);
    // }

    // this.removeSegment.value = removeText || this.defaultRemoveGroupByValue;
    // return Promise.resolve([...metricLabels, ...resourceLabels, this.removeSegment]);
    return Promise.resolve([...metricLabels, ...resourceLabels]);
  }

  async getFilterValues() {}
}
