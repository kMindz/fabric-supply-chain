import React from 'react';

export class Footer extends React.Component {
  render() {
    const now = new Date().getFullYear();
    return (
      <div className="text-center">
        <p>
          <a href="http://altoros.com" target="_blank" rel="noopener noreferrer">Altoros</a> &copy; {now}
        </p>
      </div>
    );
  }
}