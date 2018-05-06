import React from 'react';
import {render} from 'react-dom';

class App extends React.Component {

    constructor (props) {
        super(props)
        this.state = { value: "World!" };

        this.handleClick = this.handleClick.bind(this)
    }

    handleClick() {
        fetch("../endpoints/server_name", {credentials: "same-origin",})
                .then((response) => response.json())
                .then((responseJson) => {
                    this.setState( { value : responseJson.name } );
                }).catch((error) => {
                console.error(error);
            })
    }

    render () {
        return <p onClick={this.handleClick}> Hello { this.state.value } </p>;
    }
}

render(<App/>, document.getElementById('app'));
