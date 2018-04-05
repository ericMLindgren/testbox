import React, { Component } from 'react';
import { Alert, Button, Panel, PanelGroup, Grid, Row, Col } from 'react-bootstrap';
// import ChallengeDetails from './ChallengeDetails'
import './ChallengeList.css';

const ChallengeList = (props) => {
	if (props.challenges.length>0) {	
		const challengePanels = props.challenges.map(entry => {
			// entry.tags.push("other tag")
			return (
				<Panel key={entry.id} eventKey={entry.id}>
					<Panel.Heading >
						<Panel.Title toggle> {entry.name} 
								{(entry.tags && entry.tags.length>0) ?
									<span className="tag-list"> 
										<span className="glyphicon glyphicon-tags"></span> {entry.tags.join(', ')}
									</span>
								: null} 
						</Panel.Title>
					</Panel.Heading>
					<Panel.Body collapsible>
						<ChallengeDetails entry={entry} onEdit={props.onEdit} onDelete={props.onDelete}/>
						
					</Panel.Body>

				</Panel>
			)
		})
		
		return (
			<PanelGroup accordion id="challengeList">
				{challengePanels}
			</PanelGroup>
		)}

	return null
	
}



// const ChallengeDetails = ({entry}) => (
class ChallengeDetails extends Component {
	constructor(props) {
		super(props)
		this.state = {
			showWarning: false
		}
	}

	showWarning = () => {
		this.setState({showWarning:true})
	}

	hideWarning = () => {
		this.setState({showWarning:false})
	}

	render() {
		const tagList = (this.props.entry.tags ? this.props.entry.tags.join(', ') : "")
		return (
			<div className="challenge-details-panel">
				<Grid>
					<Row>
						<Col md={2}> <strong> ID: </strong></Col> 
						<Col md={5}> {this.props.entry.id} </Col>
					</Row>
					<Row>
						<Col md={2}> <strong> Summary: </strong></Col> 
						<Col md={5}> {this.props.entry.shortDesc} </Col>
					</Row>
					{this.props.entry.longDesc === "" ? null :
						<Row>
							<Col md={2}> <strong> Long Description: </strong> </Col> 
							<Col md={5}> {this.props.entry.longDesc} </Col>
						</Row>
					}
					<Row>
						<Col md={2}> <strong> Tags: </strong> </Col> 
						<Col md={5}> {"[ " + tagList + " ]"} </Col>
					</Row>
					<Row>
						<Col md={2}> <strong> Sample IO: </strong> </Col> 
						<Col md={5}> <IOPanel IO={this.props.entry.sampleIO} /> </Col>
					</Row>
					<Row>
						<Col md={2}> <strong> Test Cases: </strong> </Col> 
						<Col md={5}> <IOPanel IO={this.props.entry.cases} /> </Col>
					</Row>
				</Grid>

				{this.state.showWarning ? <DeleteWarning onAbort={this.hideWarning} onDelete={() => this.props.onDelete(this.props.entry.id)}/> : <ChallengeDetailsButtons onEdit={() => this.props.onEdit(this.props.entry.id)} onDelete={this.showWarning}/>}
			</div>
		)
	}
}

const DeleteWarning = (props) => (
	<Alert bsStyle="danger" className="top-margin-bump">
	<h4>Deleting entries is permanent!</h4>
	<p>If you delete this entry it will be permanently removed from the database, there is no undo.
	Are you sure you want to continue?</p>
	<div className="align-right">
		<Button onClick={props.onAbort}>Abort</Button> <Button bsStyle="danger" onClick={props.onDelete}>Delete</Button>
	</div>
	</Alert>
)

const ChallengeDetailsButtons = (props) => (
	<div className="align-right">
		<Button onClick={props.onEdit}>Edit</Button> <Button bsStyle="danger" onClick={props.onDelete}>Delete</Button>
	</div>
)

const IOPanel = ({IO}) => (
		<Grid className="no-padding-left">
			<Row>
				<Col md={1}> <strong> Input: </strong></Col> 
				<Col md={1}></Col>
				<Col md={1}> <strong> Expectation: </strong> </Col>
			</Row>
			{IO.map(testCase => (
				<Row key={testCase.input}>
					<Col md={1}> {testCase.input} </Col>
					<Col md={1}> <span className="glyphicon glyphicon-arrow-right"></span> </Col>
					<Col md={1}> {testCase.expect} </Col>
				</Row>
			))}
		</Grid>
)

export default ChallengeList