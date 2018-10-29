import { Modal } from 'app/types';

export enum ActionTypes {
  AddModal = 'ADD_MODAL',
  ClearModal = 'CLEAR_MODAL',
}

interface AddModal {
  type: ActionTypes.AddModal;
  payload: Modal;
}

interface ClearModalAction {
  type: ActionTypes.ClearModal;
  payload: number;
}

export type Action = AddModal | ClearModalAction;

export const clearModal = () => ({
  type: ActionTypes.ClearModal,
});

export const addModal = (modal: Modal) => ({
  type: ActionTypes.AddModal,
  payload: modal,
});
