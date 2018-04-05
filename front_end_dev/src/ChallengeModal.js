import React, { Component } from 'react';
import { FormGroup, ControlLabel, FormControl, HelpBlock, Modal, Button, Grid, Row, Col } from 'react-bootstrap';
import './ChallengeModal.css'

class ChallengeModal extends Component {
	constructor(props) {
		super(props)
		
		// this.resetState()
		const sampleIOs = {};
		const testCases = {};

		sampleIOs[newCaseID()] = {"input":"","expect":"", "desc":""}
		testCases[newCaseID()] = {"input":"","expect":"", "desc":""}

		this.state = {
			id: -1,
			sampleIOs: sampleIOs,
			testCases: testCases,
			fieldName: "",
			fieldShortDesc: "",
			fieldLongDesc: "",
			fieldTags: "",
		}
	}
	
	handleAddIO = (type) => {
		switch (type) {
			case "sample":
				const newSampleIO = Object.assign({}, this.state.sampleIOs)
				newSampleIO[newCaseID()]={"input":"","expect":"", "desc":""}
				this.setState({sampleIOs:newSampleIO})
				break

			case "test":
				const newTestCases = Object.assign({}, this.state.testCases)
				newTestCases[newCaseID()]={"input":"","expect":"", "desc":""}
				this.setState({testCases:newTestCases})
				break

		}
	}

	handleRemoveIO = (type, i) => {
		switch (type) {
			case "sample":
				const newSampleIO = Object.assign({}, this.state.sampleIOs)
				delete newSampleIO[i]
				this.setState({sampleIOs:newSampleIO})
				break

			case "test":
				const newTestCases = Object.assign({}, this.state.testCases)
				delete newTestCases[i]
				this.setState({testCases:newTestCases})
				break

		}
	}

	handleIOChange = (e) => {
		// console.log("handle", e)
		switch (e.type) {
			case "sample":
				const newSampleIO = Object.assign({}, this.state.sampleIOs)
				newSampleIO[e.key][e.field] = e.value
				this.setState({sampleIOs:newSampleIO})
				break

			case "test":
				const newTestCases = Object.assign({}, this.state.testCases)
				newTestCases[e.key][e.field] = e.value
				this.setState({testCases:newTestCases})
				break
		}
	}

	handleInputChange = (event) => {
		// alert(event.target)
	    const target = event.target;
	    const value = target.type === 'checkbox' ? target.checked : target.value;
	    const name = target.id;

	    this.setState({
	      [name]: value
	    });
	    // console.log(this.state)
	}

	resetState = () => {
		console.log('(modal) resetState')
		const sampleIOs = {};
		const testCases = {};

		sampleIOs[newCaseID()] = {"input":"","expect":""}
		testCases[newCaseID()] = {"input":"","expect":"", "desc":""}

		
		this.setState({
			id: -1,
			sampleIOs: sampleIOs,
			testCases: testCases,
			fieldName: "",
			fieldShortDesc: "",
			fieldLongDesc: "",
			fieldTags: "",
			loaded: false,
		})
	}

	loadEntry = (entry) => {
		console.log("loading entry", entry.id)

		const newSampleIO = {}
		for (let sample of entry.sampleIO) {
			newSampleIO[newCaseID()] = sample
		}

		const newTestCases = {}
		for (let testCase of entry.cases) {
			newTestCases[newCaseID()] = testCase
		}
	
		const tags = entry.tags ? entry.tags.join(', ') : ""
		console.log('loadentry props.tags', entry.tags, 'tags', tags)

		this.setState({
			id: entry.id,
			fieldName: entry.name,
			fieldShortDesc: entry.shortDesc,
			fieldLongDesc: entry.longDesc,
			fieldTags: tags,
			sampleIOs:newSampleIO,
			testCases:newTestCases,
			loaded: true,
		})
	}

	handleClose = () => {
		this.props.onHide()
	}

	handleSave = () => {
		console.log("saving entry", this.state.id)
		this.props.onSubmit(this.composeEntry())
	}

	composeEntry = () => {
		console.log('composeEntry')
		const cases = []
		const sampleIO = []

		for (let k of Object.keys(this.state.sampleIOs)) {
			sampleIO.push(this.state.sampleIOs[k])
		}

		for (let k of Object.keys(this.state.testCases)) {
			cases.push(this.state.testCases[k])
		}
		
		console.log('tag state', this.state.fieldTags)
		// const tags = this.state.fieldTags.trim() == "" ? null : this.state.fieldTags.split(',').map(s => s.trim()).filter(w => w.length > 0)
		const tags = this.state.fieldTags.split(',').map(s => s.trim()).filter(w => w.length > 0)
		
		console.log('tags after compose', tags)
		return {
			id: this.state.id,
			name: this.state.fieldName,
			shortDesc: this.state.fieldShortDesc,
			longDesc: this.state.fieldLongDesc,
			tags: tags,
			sampleIO: sampleIO,
			cases: cases
		}
	}

	componentWillReceiveProps(nextProps) {
		console.log('willRecProps')
		if (nextProps.entry) {
			this.loadEntry(nextProps.entry);
		}  else {
			this.resetState()
		}
	}

	render() {
		// console.log("modal entry: ", this.props.entry)
		return (
		   <Modal show={this.props.show} onHide={this.props.onHide}>
				<Modal.Header closeButton>
		     	   <Modal.Title>{this.props.mode} Challenge</Modal.Title>
		        </Modal.Header>

		        <Modal.Body>
					<form>
						<FieldGroup
					     	id="fieldName"
					     	type="text"
					     	label="Challenge Name:"
					     	help="A short and succinct name for this challenge"
					     	value={this.state.fieldName}
					     	onChange={this.handleInputChange}
					   	/>
					   	<FieldGroup
					     	id="fieldShortDesc"
					     	type="text"
					     	value={this.state.fieldShortDesc}
					     	onChange={this.handleInputChange}
					     	label="Short Description:"
					     	help="A minimal explanation of whats necessary to pass this challenge"
					   	/>
					   	<FieldGroup
					     	id="fieldLongDesc"
					     	type="textarea"
					     	componentClass="textarea"
					     	value={this.state.fieldLongDesc}
					     	onChange={this.handleInputChange}
					     	label="Long Description:"
					     	help="An optional, more in-depth description of challenge. Provides background or other helpful information"
					   	/>
					   	<FieldGroup
					     	id="fieldTags"
					     	type="text"
					     	value={this.state.fieldTags}
					     	onChange={this.handleInputChange}
					     	label="Tags:"
					     	help="A comma-seperated list of tags for this challenge"
					   	/>

						<IOFieldList label="Sample IO:" type="sample" showDesc={false} IOs={this.state.sampleIOs} onChange={this.handleIOChange} onRemove={this.handleRemoveIO} onAdd={this.handleAddIO}/> 
						<IOFieldList label="Test Cases:" type="test" showDesc={true} IOs={this.state.testCases} onChange={this.handleIOChange} onRemove={this.handleRemoveIO} onAdd={this.handleAddIO}/> 
					</form>
				</Modal.Body>
				<Modal.Footer>
	            	<Button bsStyle="warning" onClick={this.handleClose}>Close</Button> <Button onClick={this.handleSave}>Save</Button>
		        </Modal.Footer>
			</Modal>
		)
	}
}

const IOFieldList = (props) => {
	// console.log("IOFieldList IOs", Object.keys(props.IOs).length)
	const count = Object.keys(props.IOs).length

	return (
		<div>
		<ControlLabel>{props.label}</ControlLabel>
		{Object.keys(props.IOs).map((sampleKey, i)=>{
			const thisSample = props.IOs[sampleKey]
			// console.log("FIELDLIST sampleKey", sampleKey)
			return (
				<FormGroup key={sampleKey}>
					<IOField type={props.type} mapkey={sampleKey} entry={thisSample} showDesc={props.showDesc} onChange={props.onChange} />
					{count > 1 ? <Button className="inline" onClick={() => props.onRemove(props.type, sampleKey)}>-</Button> : null} 
					{i == count-1 ? <Button className="inline" onClick={() => props.onAdd(props.type)}>+</Button> : null}
				</FormGroup>

			)
		})}
		</div>
	)
}




const IOField = ({mapkey, entry, showDesc, onChange, type}) => {
		// console.log("IOFIELD type", entry)
		return (
			<div className="inline">

				<FormControl className="margin-right-bump width-30-percent inline vert-align-top" componentClass="textarea"  id="input" value={entry.input} onChange={(event) => onChange({key:mapkey, field:"input", value:event.target.value, type:type})}  />
				<span className="glyphicon glyphicon-arrow-right"></span>
				<FormControl className="margin-right-bump width-30-percent inline vert-align-top" componentClass="textarea"  id="expect" value={entry.expect} onChange={(event) => onChange({key:mapkey, field:"expect", value:event.target.value, type:type})}  />
				{showDesc ? <FormControl className="margin-right-bump width-30-percent inline" componentClass="textarea"  id="desc" value={entry.desc} onChange={(event) => onChange({key:mapkey, field:"desc", value:event.target.value, type:type})} /> : null}
			</div>
			)
}


const FieldGroup = ({id, label, help, ...props }) => {
  return (
    <FormGroup controlId={id}>
      <ControlLabel>{label}</ControlLabel>
      <FormControl {...props} />
      {help && <HelpBlock>{help}</HelpBlock>}
    </FormGroup>
  );
}

let caseID = -1
const newCaseID = () => {
	caseID ++
	return caseID.toString()
}

export default ChallengeModal