import { DashboardShareModal } from './variants/DashboardShareModal';
import { RemoveModal } from './variants/RemoveModal';

export enum ModalVariants {
  DashboardShareModal = 'DASHBOARD_SHARE_MODAL',
  RemoveModal = 'REMOVE_MODAL', //{/* aka ConfirmModal? */ }
}

export const ModalData = {
  [ModalVariants.DashboardShareModal]: {
    component: DashboardShareModal,
  },
  [ModalVariants.RemoveModal]: {
    component: RemoveModal,
  },
};
