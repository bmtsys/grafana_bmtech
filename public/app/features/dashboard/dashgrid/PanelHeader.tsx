import React from 'react';
import classNames from 'classnames';
import { PanelModel } from '../panel_model';
import { DashboardModel } from '../dashboard_model';
import { store } from 'app/store/configureStore';
import { updateLocation } from 'app/core/actions';
import { StoreState } from 'app/types';
import { connectWithStore } from 'app/core/utils/connectWithReduxStore';
import { addModal, clearModal } from 'app/core/actions';
import ModalPortal from 'app/core/components/Modals/ModalPortal';
import { SampleModalContentForPOC } from './SampleModalContentForPOC';

interface PanelHeaderProps {
  panel: PanelModel;
  dashboard: DashboardModel;
  addModal: typeof addModal;
  clearModal: typeof clearModal;
  isMyModal: boolean;
}

const ModalName = 'POC_MODAL';

export class PanelHeader extends React.Component<PanelHeaderProps, any> {
  onEditPanel = () => {
    store.dispatch(
      updateLocation({
        query: {
          panelId: this.props.panel.id,
          edit: true,
          fullscreen: true,
        },
      })
    );
  };

  onViewPanel = () => {
    store.dispatch(
      updateLocation({
        query: {
          panelId: this.props.panel.id,
          edit: false,
          fullscreen: true,
        },
      })
    );
  };

  onAddModal = () => {
    this.props.addModal({
      name: ModalName,
      attributes: {
        onClose: () => {},
      },
    });
  };

  onCloseModal = () => {
    this.props.clearModal();
  };

  render() {
    const isFullscreen = false;
    const isLoading = false;
    const panelHeaderClass = classNames({ 'panel-header': true, 'grid-drag-handle': !isFullscreen });
    const { isMyModal } = this.props;

    return [
      <div className={panelHeaderClass}>
        <span className="panel-info-corner">
          <i className="fa" />
          <span className="panel-info-corner-inner" />
        </span>

        {isLoading && (
          <span className="panel-loading">
            <i className="fa fa-spinner fa-spin" />
          </span>
        )}

        <div className="panel-title-container">
          <button className="btn btn-primary" onClick={this.onAddModal}>
            Open modal POC
          </button>
          <span className="panel-title">
            <span className="icon-gf panel-alert-icon" />
            <span className="panel-title-text">{this.props.panel.title}</span>
            <span className="panel-menu-container dropdown">
              <span className="fa fa-caret-down panel-menu-toggle" data-toggle="dropdown" />
              <ul className="dropdown-menu dropdown-menu--menu panel-menu" role="menu">
                <li>
                  <a onClick={this.onEditPanel}>
                    <i className="fa fa-fw fa-edit" /> Edit
                  </a>
                </li>
                <li>
                  <a onClick={this.onViewPanel}>
                    <i className="fa fa-fw fa-eye" /> View
                  </a>
                </li>
              </ul>
            </span>
            <span className="panel-time-info">
              <i className="fa fa-clock-o" /> 4m
            </span>
          </span>
        </div>
      </div>,
      isMyModal ? (
        <ModalPortal onCloseModal={this.onCloseModal} title="Modal title" icon="fa fa-share-square-o">
          <SampleModalContentForPOC /> {/* This could be something like <ShareDashboardModal /> or similar */}
        </ModalPortal>
      ) : null,
    ];
  }
}

const mapStateToProps = (state: StoreState) => {
  const modals = state.modals.modals;

  return {
    modals: modals,
    currModal: modals.length > 0 ? modals[modals.length - 1] : null,
    isMyModal: modals.length > 0 ? modals[modals.length - 1].name === ModalName : false,
  };
};

const mapDispatchToProps = {
  clearModal,
  addModal,
};

export default connectWithStore(PanelHeader, mapStateToProps, mapDispatchToProps);
