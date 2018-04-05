import React, { Component } from 'react';
import axios from 'axios';

import ChallengeList from './ChallengeList'
import ChallengeModal from './ChallengeModal'
import './App.css';

// Bootstrap imports
import { Panel, Button, Col, Form, FormControl, FormGroup } from 'react-bootstrap';

class App extends Component {
  constructor(props) {
    super(props)
    this.state = {
      searchResults: [],
      show: false,
      focusEntry: null,
      mode: null,
      reset: false
    }
  }

  populateChallenges = () => {
    // should query all challenges and hold in memory at what point does TODO
    // the inefficiency of this become prohibative...?

    axios.get('http://localhost:31336/challenges/all/')
         .then(res => {
            console.log(res.data.result)
            const corpus = JSON.parse(res.data.result)
            console.table(corpus)
            this.setState({
              corpus: corpus,
              // searchResults: corpus
            })
            this.handleSearchChange()
         })

  }


  handleSearchChange = (event) => {
    const searchString = event ? event.target.value.toLowerCase() : ""

    let results = []
    for (let chal of Object.values(this.state.corpus)) {
      // if (event.target.value == chal.name)
      if (chal.name.toLowerCase().match(searchString)){
        results.push(chal)
        console.log("match")
      }

    }
    results.sort()
    this.setState({
      searchResults: results
    })
  }

  handleChallengeSubmit = (entry) => {
    console.log("top level submit", entry)
    const entryJSON = JSON.stringify(entry)
    let action;
    switch (entry.id) {
      case -1:
        action = "insert"
        break
      default:
        action = "update"
        //submit update challenge!
    }

    axios.put(
            "http://localhost:31336/challenges/" + action, 
            entryJSON, 
            {headers: {"Content-Type": "application/json"}}
        )
        .then(r => {
          alert("Submit response:", r.status);
          this.populateChallenges();
          this.handleHide();
          console.log("DONE!")
        })
        .catch(e => console.log(e));
  }

  handleNew = () => {
    this.setState({
      mode: "New",
      show:true,
      focusEntry: null,
      reset:true
    })
  }

  handleEdit = (id) => {
    console.log('editing', id)
    console.log('corpus', this.state.corpus)
    this.setState({
      mode: "Edit",
      show:true,
      focusEntry: this.state.corpus[id]
    })
  }

  handleHide = () => {
    console.log('master hide')
    this.setState({
      show: false,
      focusEntry: null
    })
  }

  handleDelete = (id) => {
    console.log('deleting challneg' + id)
    // const entryJSON = JSON.stringify(id)
    axios.put(
            "http://localhost:31336/challenges/delete", 
            id, 
            {headers: {"Content-Type": "application/json"}}
        )
        .then(r => {
          console.log(r.status);
          // this.handleHide();
          this.populateChallenges();
          console.log("DELETE DONE!")
        })
        .catch(e => console.log(e));
  }

  handleModalReset = () => {
    this.setState({reset:false})
  }

  componentDidMount() {
    this.populateChallenges()
  }

  render() {
    
    // if (this.state.searchResults)
    //   var someList = this.state.searchResults.map(data => data.description)

    return (
      <div className="App">
        <header className="App-header">
         
          <h1 className="App-title">Challenge Librarian</h1>
        </header>

        <div className="main-panel">
          <Panel>
          <Panel.Body>
            <Form horizontal>
              <FormGroup>
                <Col sm={3}>
                  <FormControl type="text" placeholer="Search Challenges..." onChange={this.handleSearchChange}/>
                </Col>
                <Col sm={1}>
                  <Button bsStyle="primary" onClick={() => this.handleNew()}>+</Button>
                </Col>
              </FormGroup>
            </Form>
              <ChallengeList challenges={this.state.searchResults} onEdit={this.handleEdit} onHide={this.handleHide} onDelete={this.handleDelete} onSubmit={this.handleChallengeSubmit}/>
          </Panel.Body>
          </Panel>
        </div>
        <ChallengeModal onSubmit={this.handleChallengeSubmit} show={this.state.show} mode={this.state.mode} reset={this.state.reset} onReset={this.handleModalReset} onHide={this.handleHide} entry={this.state.focusEntry}/>
      </div>  
    );
  }
}


export default App;
