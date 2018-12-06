import React, { PureComponent } from 'react';
import Select from 'react-select';
import ResetStyles from '../../core/components/Picker/ResetStyles';
import PickerOption from '../../core/components/Picker/PickerOption';
import IndicatorsContainer from '../../core/components/Picker/IndicatorsContainer';
import NoOptionsMessage from '../../core/components/Picker/NoOptionsMessage';
import TimePicker from './TimePicker';
import { ExploreDatasource } from 'app/types/explore';
import { RawTimeRange } from 'app/types/series';

interface Props {
  dataSourceMissing: boolean;
  dataSourceLoading: boolean;
  exploreDataSources: ExploreDatasource[];
  loading: boolean;
  onChangeDataSource: (origin) => void;
  onChangeTime: (nextRange, scanning?) => void;
  onClear: () => void;
  onCloseSplit: () => void;
  onSubmit: () => void;
  position: string;
  range: RawTimeRange;
  selectedDataSource: ExploreDatasource;
  setTimePickerRef: (element) => void;
  split: boolean;
}

export default class ExploreNavBar extends PureComponent<Props> {
  render() {
    const {
      dataSourceMissing,
      dataSourceLoading,
      exploreDataSources,
      loading,
      onCloseSplit,
      onChangeDataSource,
      onChangeTime,
      onClear,
      onSubmit,
      position,
      range,
      selectedDataSource,
      setTimePickerRef,
      split,
    } = this.props;

    return (
      <div className="navbar">
        {position === 'left' ? (
          <div>
            <a className="navbar-page-btn">
              <i className="fa fa-rocket" />
              Explore
            </a>
          </div>
        ) : (
          <div className="navbar-buttons explore-first-button">
            <button className="btn navbar-button" onClick={onCloseSplit}>
              Close Split
            </button>
          </div>
        )}
        {!dataSourceMissing && (
          <div className="navbar-buttons">
            <Select
              classNamePrefix={`gf-form-select-box`}
              isMulti={false}
              isLoading={dataSourceLoading}
              isClearable={false}
              className="gf-form-input gf-form-input--form-dropdown datasource-picker"
              onChange={onChangeDataSource}
              options={exploreDataSources}
              styles={ResetStyles}
              placeholder="Select datasource"
              loadingMessage={() => 'Loading datasources...'}
              noOptionsMessage={() => 'No datasources found'}
              value={selectedDataSource}
              components={{
                Option: PickerOption,
                IndicatorsContainer,
                NoOptionsMessage,
              }}
            />
          </div>
        )}
        <div className="navbar__spacer" />
        {position === 'left' &&
          !split && (
            <div className="navbar-buttons">
              <button className="btn navbar-button" onClick={onCloseSplit}>
                Split
              </button>
            </div>
          )}
        <TimePicker ref={element => setTimePickerRef(element)} range={range} onChangeTime={onChangeTime} />
        <div className="navbar-buttons">
          <button className="btn navbar-button navbar-button--no-icon" onClick={onClear}>
            Clear All
          </button>
        </div>
        <div className="navbar-buttons relative">
          <button className="btn navbar-button--primary" onClick={onSubmit}>
            Run Query{' '}
            {loading ? <i className="fa fa-spinner fa-spin run-icon" /> : <i className="fa fa-level-down run-icon" />}
          </button>
        </div>
      </div>
    );
  }
}
