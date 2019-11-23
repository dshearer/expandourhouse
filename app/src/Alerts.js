import React, { Component } from 'react';
import PropTypes from 'prop-types';

class ShowSuccess extends Component {
  constructor(props) {
    super(props);

    this.state = {
      message: undefined,
      visible: true,
    };
  }

  show(message) {
    if (this.timer) {
      clearInterval(this.timer);
    }

    this.setState({
      message: message,
      visible: true,
    });
    this.timer = setInterval(
      () => {this.setState({visible: false});},
      10000
    );
  }

  componentWillUnmount() {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = undefined;
    }
  }

  render() {
    if (!this.state.visible || !this.state.message) {
      return null;
    }

    return (
      <div className="alert alert-success" role="alert">
        {this.state.message}
      </div>
    );
  }
}

const showErrorsPropTypes = {
  errors: PropTypes.array,
};

class ShowErrors extends Component {
  render() {
    const errors = this.props.errors ? this.props.errors : [];

    var contents;
    if (errors.length === 0) {
      return null;
    }
    else if (errors.length === 1) {
      contents = errors[0];
    }
    else {
      const items = [];
      for (var i = 0; i < errors.length; ++i) {
        items.push(<li key={i}>{errors[i]}</li>);
      }
      contents = (<ul className="pl-0 mb-0">{items}</ul>);
    }

    return (
      <div className="alert alert-danger" role="alert">
        {contents}
      </div>
    );
  }
}

ShowErrors.propTypes = showErrorsPropTypes;

function getAxiosErrors(err) {
  if (err.response) {
    const data = err.response.data;
    if (!data) {
      return ['Unknown error'];
    }
    else if (typeof data === 'object' && data.jsonSchemaValidation) {
      const errors = [];
      const validations = data.validations.body;
      for (var i = 0; i < validations.length; ++i) {
        const v = validations[i];
        const prop = v.property.split('.')[2];
        const msg = `${prop} ${v.messages[0]}`;
        errors.push(msg);
      }
      return errors;
    }
    else {
      return ['' + err.response.data];
    }
  }
  else if (err.request) {
    /* No response was received. */
    return ['Backend timeout'];
  }
  else {
    return [err.message];
  }
}

export {ShowSuccess, ShowErrors, getAxiosErrors};
