import React, { PureComponent } from 'react';
import { hot } from 'react-hot-loader';
import { connect } from 'react-redux';
import PageHeader from 'app/core/components/PageHeader/PageHeader';
import { getNavModel } from 'app/core/selectors/navModel';
import { NavModel, StoreState, FolderState } from 'app/types';
import { getLoadingNav } from './state/navModel';
import { getFolderByUid } from './state/actions';

export interface Props {
  navModel: NavModel;
  folderUid: string;
  folder: FolderState;
  getFolderByUid: typeof getFolderByUid;
}

export class FolderDashboardsPage extends PureComponent<Props> {
  componentDidMount() {
    this.props.getFolderByUid(this.props.folderUid);
  }

  render() {
    const { navModel } = this.props;

    return (
      <div>
        <PageHeader model={navModel} />
      </div>
    );
  }
}

const mapStateToProps = (state: StoreState) => {
  const uid = state.location.routeParams.uid;

  return {
    navModel: getNavModel(state.navIndex, `folder-dashboards-${uid}`, getLoadingNav(2)),
    folderUid: uid,
    folder: state.folder,
  };
};

const mapDispatchToProps = {
  getFolderByUid,
};

export default hot(module)(connect(mapStateToProps, mapDispatchToProps)(FolderDashboardsPage));
