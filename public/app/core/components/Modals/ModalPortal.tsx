import React, { SFC } from 'react';
import ReactDOM from 'react-dom';

interface Props {
  title: string;
  icon: string;
  children: any;
  onCloseModal: () => void;
}

const ModalPortal: SFC<Props> = props => {
  const { title, icon, children, onCloseModal } = props;

  return ReactDOM.createPortal(
    <div className="modal">
      <div className="modal-body">
        <div className="modal-header">
          <h2 className="modal-header-title">
            {icon && <i className={icon} />}
            <span className="p-l-1">{title}</span>
          </h2>
          <a className="modal-header-close" onClick={onCloseModal}>
            <i className="fa fa-remove" />
          </a>
        </div>
        <div className="modal-content">
          {children}
          <button className="btn btn-primary" onClick={onCloseModal}>
            Close
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
};

export default ModalPortal;
