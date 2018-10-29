import { Modal, ModalsState } from 'app/types/';
import { Action, ActionTypes } from 'app/core/actions/modal';
export const initialState: ModalsState = {
  modals: [] as Modal[],
};

export const modalsReducer = (state = initialState, action: Action): ModalsState => {
  switch (action.type) {
    case ActionTypes.AddModal:
      return {
        ...state,
        modals: state.modals.concat([action.payload]),
      };
    case ActionTypes.ClearModal:
      return {
        ...state,
        modals: state.modals.slice(0, -1),
      };
  }
  return state;
};
