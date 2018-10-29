import React, { PureComponent } from 'react';
import { connectWithStore } from 'app/core/utils/connectWithReduxStore';
import { clearModal } from 'app/core/actions';
import { ModalData } from './ModalVariants';
import { StoreState } from 'app/types';

export interface Props {
  modals: any;
  currModal: any;
  clearModal: typeof clearModal;
}

export class Modal extends PureComponent<Props> {
  onCloseClick = () => {
    const { clearModal } = this.props;
    clearModal();
  };

  render() {
    const { modals, currModal } = this.props;
    if (!currModal) {
      return null;
    }
    const attributes = currModal.attributes;
    const ComponentToRender = ModalData[currModal.variant].component;
    return modals.length > 0 ? (
      <div className="modal">
        <div className="modal-body">
          <div className="modal-header">
            <h2 className="modal-header-title">
              {attributes.icon && <i className={attributes.icon} />}
              <span className="p-l-1">{attributes.title}</span>
            </h2>
            <a className="modal-header-close" onClick={attributes.onCloseModal}>
              <i className="fa fa-remove" />
            </a>
          </div>
          <div className="modal-content">
            <ComponentToRender {...attributes} />
            <button className="btn btn-primary" onClick={attributes.onCloseModal}>
              Close
            </button>
          </div>
        </div>
      </div>
    ) : null;
  }
}

const mapStateToProps = (state: StoreState) => ({
  modals: state.modals.modals,
  currModal: state.modals.modals.length > 0 ? state.modals.modals[state.modals.modals.length - 1] : null,
});

const mapDispatchToProps = {
  clearModal,
};

export default connectWithStore(Modal, mapStateToProps, mapDispatchToProps);
