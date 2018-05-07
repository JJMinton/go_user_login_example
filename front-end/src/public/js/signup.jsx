import React from 'react';
import {render} from 'react-dom';

class App extends React.Component {

    constructor (props) {
        super(props)
        this.state = { value: "username",
                       response: "" };

        this.handleSubmit = this.handleSubmit.bind(this);
        this.handleChange = this.handleChange.bind(this);
    }

    handleChange(event) {
        this.setState( {value: event.target.value } );
    }

    handleSubmit(event) {
        event.preventDefault();
        fetch("endpoints/user", {method: "POST",
                                 headers: {'Content-Type':'application/json'},
                                 body: JSON.stringify({'username': this.state.value }),
                                 credentials: "same-origin"})
                .then((response) => {
                    if (response == "username taken") {
                        this.state.response = "That username is taken";
                    } else if (response == "error" || response == "") {
                        this.state.response = "System error. Try again.";
                    } else {
                        window.location.href = "/auth/google/login";
                    }
                }).catch((error) => {
                console.error(error);
            });
    }

    render() {
        return (
        <form onSubmit={this.handleSubmit}>
        Username: <input type="text" name="Username" value={ this.state.value } onChange={this.handleChange}/><br/>
        { this.state.response ? this.state.response + <br/> : '' }
        <input type="submit" value="Sign up"/>
        </form>
        )
    }
}

render(<App/>, document.getElementById('app'));
