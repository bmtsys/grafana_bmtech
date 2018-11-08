import React, { PureComponent } from 'react';
import { FadeIn } from 'app/core/components/Animations/FadeIn';

interface Props {
  nr: string;
  title: string;
  selectedText?: string;
  selectedImage?: string;
  children: any;
  onToggleSelect?: () => void;
}

interface State {
  expanded: boolean;
}

export class EditSection extends PureComponent<Props, State> {
  constructor(props) {
    super(props);

    this.state = {
      expanded: true,
    };
  }

  onToggleExpand = () => {
    this.setState({ expanded: !this.state.expanded });
  };

  onToggleSelect = event => {
    event.stopPropagation();
    this.props.onToggleSelect();
  };

  render() {
    const { nr, title, selectedText, selectedImage, children } = this.props;

    return (
      <div className="edit-section">
        <div className="edit-section__header" onClick={this.onToggleExpand}>
          <div className="edit-section__nr">{nr}</div>
          <div className="edit-section__title">{title}</div>
          {selectedText && (
            <div className="edit-section__selected" onClick={this.onToggleSelect}>
              <img className="edit-section__selected-image" src={selectedImage} />
              <div className="edit-section__selected-name">{selectedText}</div>
              <i className="fa fa-caret-down" />
            </div>
          )}
          <div className="edit-section__expand-state">
            <i className="fa fa-chevron-down" />
          </div>
        </div>

        <FadeIn in={this.state.expanded} duration={200}>
          <div>
            <div className="edit-section__line" />
            <div className="edit-section__body">{children}</div>
          </div>
        </FadeIn>
      </div>
    );
  }
}
