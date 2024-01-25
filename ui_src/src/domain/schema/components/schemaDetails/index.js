// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { AddRounded, CheckCircleOutlineRounded, ErrorOutlineRounded } from '@material-ui/icons';
import Editor, { DiffEditor, loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import React, { useContext, useEffect, useState } from 'react';
import Schema from 'protocol-buffers-schema';
import { getUnique, isThereDiff, parsingDate } from 'services/valueConvertor';
import { ReactComponent as StationsActiveIcon } from 'assets/images/stationsIconActive.svg';
import { ReactComponent as CreatedDateIcon } from 'assets/images/createdDateIcon.svg';
import { ReactComponent as ScrollBackIcon } from 'assets/images/scrollBackIcon.svg';
import { ReactComponent as RedirectIcon } from 'assets/images/redirectIcon.svg';
import { ReactComponent as CreatedByIcon } from 'assets/images/createdByIcon.svg';
import { ReactComponent as VerifiedIcon } from 'assets/images/verifiedIcon.svg';
import { ReactComponent as RollBackIcon } from 'assets/images/rollBackIcon.svg';
import SelectVersion from 'components/selectVersion';
import { ReactComponent as TypeIcon } from 'assets/images/typeIcon.svg';
import { ApiEndpoints } from 'const/apiEndpoints';
import SelectComponent from 'components/select';
import { httpRequest } from 'services/http';
import { isCloud } from 'services/valueConvertor';
import Button from 'components/button';
import Modal from 'components/modal';
import Copy from 'components/copy';
import TagsList from 'components/tagList';
import LockFeature from 'components/lockFeature';
import { useHistory } from 'react-router-dom';
import pathDomains from 'router';
import { Context } from 'hooks/store';
import Ajv2019 from 'ajv/dist/2019';
import jsonSchemaDraft04 from 'ajv-draft-04';
import draft7MetaSchema from 'ajv/dist/refs/json-schema-draft-07.json';
import Ajv2020 from 'ajv/dist/2020';
import draft6MetaSchema from 'ajv/dist/refs/json-schema-draft-06.json';
import OverflowTip from 'components/tooltip/overflowtip';
import { validate, parse, buildASTSchema } from 'graphql';
import SegmentButton from 'components/segmentButton';
import AttachStationModal from '../attachStationModal';
import { showMessages } from 'services/genericServices';
const avro = require('avro-js');

loader.init();
loader.config({ monaco });

function SchemaDetails({ schemaName, closeDrawer }) {
    const ajv = new Ajv2019();
    const [state, dispatch] = useContext(Context);

    const [versionSelected, setVersionSelected] = useState();
    const [currentVersion, setCurrentversion] = useState();
    const [updated, setUpdated] = useState(false);
    const [loading, setIsLoading] = useState(false);
    const [attachLoader, setAttachLoader] = useState(false);
    const [rollLoading, setIsRollLoading] = useState(false);
    const [newVersion, setNewVersion] = useState('');
    const [schemaDetails, setSchemaDetails] = useState({
        schema_name: '',
        type: '',
        version: [],
        tags: [],
        used_stations: []
    });
    const [rollBackModal, setRollBackModal] = useState(false);
    const [activateVersionModal, setActivateVersionModal] = useState(false);
    const [isDiff, setIsDiff] = useState('Yes');
    const [validateLoading, setValidateLoading] = useState(false);
    const [validateError, setValidateError] = useState('');
    const [validateSuccess, setValidateSuccess] = useState(false);
    const [messageStructName, setMessageStructName] = useState('');
    const [messagesStructNameList, setMessagesStructNameList] = useState([]);
    const [editable, setEditable] = useState(false);
    const [attachStaionModal, setAttachStaionModal] = useState(false);
    const [isUpdate, setIsUpdate] = useState(false);
    const [latestVersion, setLatest] = useState({});

    const history = useHistory();

    const goToStation = (stationName) => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };

    const arrangeData = (schema) => {
        let index = schema.versions?.findIndex((version) => version?.active === true);
        setCurrentversion(schema.versions[index]);
        setVersionSelected(schema.versions[index]);
        setNewVersion(schema.versions[index].schema_content);
        setSchemaDetails(schema);
        if (schema.type === 'protobuf') {
            let parser = Schema.parse(schema.versions[index].schema_content).messages;
            setMessageStructName(schema.versions[index].message_struct_name);
            if (parser.length === 1) {
                setEditable(false);
            } else {
                setEditable(true);
                setMessageStructName(schema.versions[index].message_struct_name);
                setMessagesStructNameList(parser);
            }
        }
    };

    const getScemaDetails = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_SCHEMA_DETAILS}?schema_name=${schemaName}`);
            arrangeData(data);
        } catch (err) {}
    };

    useEffect(() => {
        getScemaDetails();
        return () => {
            setValidateSuccess('');
            setValidateError('');
        };
    }, []);

    const handleSelectVersion = (e) => {
        let index = schemaDetails?.versions?.findIndex((version) => version.version_number === e);
        setVersionSelected(schemaDetails?.versions[index]);
        setMessageStructName(schemaDetails?.versions[index].message_struct_name);
        setNewVersion('');
    };

    const createNewVersion = async () => {
        try {
            setIsLoading(true);
            const data = await httpRequest('POST', ApiEndpoints.CREATE_NEW_VERSION, {
                schema_name: schemaName,
                schema_content: newVersion,
                message_struct_name: messageStructName
            });
            if (data) {
                arrangeData(data);
                setLatest(data);
                setActivateVersionModal(true);
                setIsLoading(false);
            }
        } catch (err) {
            if (err.status === 555) {
                setValidateSuccess('');
                setValidateError(err.data.message);
            }
        }
        setIsLoading(false);
    };

    const rollBackVersion = async (latest = false) => {
        try {
            setIsRollLoading(true);
            const data = await httpRequest('PUT', ApiEndpoints.ROLL_BACK_VERSION, {
                schema_name: schemaName,
                version_number: latest ? latestVersion?.versions[0]?.version_number : versionSelected?.version_number
            });
            if (data) {
                arrangeData(data);
                showMessages('success', 'Your selected version is now the primary version');
                setRollBackModal(false);
                setActivateVersionModal(false);
                if (schemaDetails?.used_stations?.length > 0) {
                    setIsUpdate(true);
                    setAttachStaionModal(true);
                }
            }
        } catch (err) {}
        setIsRollLoading(false);
    };

    const validateJsonSchemaContent = (value, ajv) => {
        const isValid = ajv.validateSchema(value);
        if (isValid) {
            setValidateSuccess('');
            setValidateError('');
        } else {
            setValidateError('Your schema is invalid');
        }
    };

    const validateProtobufSchema = (value) => {
        try {
            let parser = Schema.parse(value).messages;
            if (parser.length > 1) {
                setEditable(true);
                setMessagesStructNameList(getUnique(parser));
            } else {
                setMessageStructName(parser[0].name);
                setEditable(false);
            }
            setValidateSuccess('');
            setValidateError('');
        } catch (error) {
            setValidateSuccess('');
            setValidateError(error.message);
        }
    };

    const validateGraphQlSchema = (value) => {
        try {
            var documentNode = parse(value);
            var graphqlSchema = buildASTSchema(documentNode);
            validate(graphqlSchema, documentNode);
            if (documentNode.definitions.length > 1) {
                setEditable(true);
                let list = [];
                documentNode.definitions.map((def) => {
                    list.push(def.name.value);
                });
                setMessagesStructNameList(list);
            } else {
                setMessageStructName(documentNode.definitions[0].name.value);
                setEditable(false);
            }
            setValidateSuccess('');
            setValidateError('');
        } catch (error) {
            setValidateSuccess('');
            setValidateError(error.message);
        }
    };

    const validateJsonSchema = (value) => {
        try {
            value = JSON.parse(value);
            ajv.addMetaSchema(draft7MetaSchema);
            validateJsonSchemaContent(value, ajv);
        } catch (error) {
            try {
                const ajv = new jsonSchemaDraft04();
                validateJsonSchemaContent(value, ajv);
            } catch (error) {
                try {
                    const ajv = new Ajv2020();
                    validateJsonSchemaContent(value, ajv);
                } catch (error) {
                    try {
                        ajv.addMetaSchema(draft6MetaSchema);
                        validateJsonSchemaContent(value, ajv);
                    } catch (error) {
                        setValidateSuccess('');
                        setValidateError(error.message);
                    }
                }
            }
        }
    };

    const validateAvroSchema = (value) => {
        try {
            avro.parse(value);
            setValidateSuccess('');
            setValidateError('');
        } catch (error) {
            setValidateSuccess('');
            setValidateError('Your schema is invalid');
        }
    };

    const checkContent = (value) => {
        const { type } = schemaDetails;
        if (value === ' ' || value === '') {
            setValidateSuccess('');
            setValidateError('Schema content cannot be empty');
        }
        if (value && value.length > 0) {
            if (type === 'protobuf') {
                validateProtobufSchema(value);
            } else if (type === 'json') {
                validateJsonSchema(value);
            } else if (type === 'graphql') {
                validateGraphQlSchema(value);
            } else if (type === 'avro') {
                validateAvroSchema(value);
            }
        }
    };

    const handleEditVersion = (value) => {
        setValidateSuccess('');
        setNewVersion(value);
        setUpdated(isThereDiff(versionSelected?.schema_content, value));
        checkContent(value);
    };

    const handleValidateSchema = async () => {
        setValidateLoading(true);
        try {
            const data = await httpRequest('POST', ApiEndpoints.VALIDATE_SCHEMA, {
                schema_type: schemaDetails?.type,
                schema_content: newVersion || versionSelected?.schema_content
            });
            if (data.is_valid) {
                setValidateError('');
                setTimeout(() => {
                    setValidateSuccess('Schema is valid');
                    setValidateLoading(false);
                }, 1000);
            }
        } catch (error) {
            if (error.status === 555) {
                setValidateSuccess('');
                setValidateError(error.data.message);
            }
            setValidateLoading(false);
        }
    };

    const removeTag = async (tagName) => {
        try {
            await httpRequest('DELETE', `${ApiEndpoints.REMOVE_TAG}`, { name: tagName, entity_type: 'schema', entity_name: schemaName });
            let tags = schemaDetails?.tags;
            let updatedTags = tags.filter((tag) => tag.name !== tagName);
            updateTags(updatedTags);
            dispatch({ type: 'SET_SCHEMA_TAGS', payload: { schemaName: schemaName, tags: updatedTags } });
        } catch (error) {}
    };

    const updateTags = (newTags) => {
        let updatedValue = { ...schemaDetails };
        updatedValue['tags'] = newTags;
        setSchemaDetails((schemaDetails) => ({ ...schemaDetails, ...updatedValue }));
        dispatch({ type: 'SET_SCHEMA_TAGS', payload: { schemaName: schemaName, tags: newTags } });
    };

    const updateStations = (stationsList) => {
        let updatedValue = { ...schemaDetails };
        updatedValue['used_stations'] = [...updatedValue['used_stations'], ...stationsList];
        setSchemaDetails((schemaDetails) => ({ ...schemaDetails, ...updatedValue }));
        dispatch({ type: 'SET_IS_USED', payload: { schemaName: schemaName } });
    };

    return (
        <schema-details is="3xd">
            <div className="scrollable-wrapper">
                <div className="type-created">
                    <div className="wrapper">
                        <TypeIcon alt="typeIcon" />
                        <p>Type:</p>
                        {schemaDetails?.type === 'json' ? <span className="capitalize">JSON schema</span> : <span className="capitalize"> {schemaDetails?.type}</span>}
                    </div>
                    <div className="wrapper">
                        <CreatedByIcon alt="createdByIcon" />
                        <p>Created by:</p>
                        <OverflowTip text={currentVersion?.created_by_username} maxWidth={'150px'}>
                            <span>{currentVersion?.created_by_username}</span>
                        </OverflowTip>
                    </div>
                    <div className="wrapper">
                        <CreatedDateIcon alt="createdDateIcon" />
                        <span>{parsingDate(currentVersion?.created_at)}</span>
                    </div>
                </div>
                <div className="tags">
                    <TagsList
                        tagsToShow={5}
                        className="tags-list"
                        tags={schemaDetails?.tags}
                        addNew={true}
                        editable={true}
                        handleDelete={(tag) => removeTag(tag)}
                        entityType={'schema'}
                        entityName={schemaName}
                        handleTagsUpdate={(tags) => {
                            updateTags(tags);
                        }}
                    />
                </div>
                <div className="schema-fields">
                    <div className="left">
                        <p className={!versionSelected?.active ? 'tlt seperator' : 'tlt'}>Schema editor</p>
                        {!versionSelected?.active && (
                            <>
                                <span>Diff : </span>
                                <SegmentButton options={['Yes', 'No']} onChange={(e) => setIsDiff(e)} />
                            </>
                        )}
                        {/* <RadioButton options={formatOption} radioValue={passwordType} onChange={(e) => passwordTypeChange(e)} /> */}
                    </div>
                    <SelectVersion value={versionSelected?.version_number} options={schemaDetails?.versions} onChange={(e) => handleSelectVersion(e)} />
                </div>
                <div className="schema-content">
                    <div className="header">
                        <div className="structure-message">
                            {schemaDetails.type === 'protobuf' && (
                                <>
                                    <p className="field-name">Master message :</p>
                                    <SelectComponent
                                        value={messageStructName}
                                        colorType="black"
                                        backgroundColorType="white"
                                        borderColorType="gray-light"
                                        radiusType="semi-round"
                                        minWidth="12vw"
                                        width="250px"
                                        height="30px"
                                        options={messagesStructNameList}
                                        iconColor="gray"
                                        popupClassName="message-option"
                                        onChange={(e) => {
                                            setMessageStructName(e);
                                            setUpdated(true);
                                        }}
                                        disabled={!editable}
                                    />
                                </>
                            )}
                        </div>
                        <div className="validation">
                            <Button
                                width="100px"
                                height="28px"
                                placeholder={
                                    <div className="validate-placeholder">
                                        <VerifiedIcon alt="verifiedIcon" />
                                        <p>Validate</p>
                                    </div>
                                }
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontFamily="InterMedium"
                                disabled={updated && newVersion === ''}
                                isLoading={validateLoading}
                                onClick={() => handleValidateSchema()}
                            />
                        </div>
                        <div className="copy-icon">
                            <Copy data={newVersion || versionSelected?.schema_content} />
                        </div>
                    </div>
                    {versionSelected?.active && (
                        <Editor
                            options={{
                                minimap: { enabled: false },
                                scrollbar: { verticalScrollbarSize: 3 },
                                scrollBeyondLastLine: false,
                                roundedSelection: false,
                                formatOnPaste: true,
                                formatOnType: true,
                                fontSize: '14px',
                                fontFamily: 'Inter'
                            }}
                            language={schemaDetails?.type === 'protobuf' ? 'proto' : schemaDetails?.type === 'avro' ? 'json' : schemaDetails?.type}
                            height="calc(100% - 55px)"
                            defaultValue={versionSelected?.schema_content}
                            value={newVersion}
                            onChange={(value) => {
                                handleEditVersion(value);
                            }}
                        />
                    )}
                    {!versionSelected?.active && (
                        <>
                            {isDiff === 'No' && (
                                <Editor
                                    options={{
                                        minimap: { enabled: false },
                                        scrollbar: { verticalScrollbarSize: 3 },
                                        scrollBeyondLastLine: false,
                                        roundedSelection: false,
                                        formatOnPaste: true,
                                        formatOnType: true,
                                        readOnly: true,
                                        fontSize: '14px',
                                        fontFamily: 'Inter'
                                    }}
                                    language={schemaDetails?.type === 'protobuf' ? 'proto' : schemaDetails?.type === 'avro' ? 'json' : schemaDetails?.type}
                                    height="calc(100% - 100px)"
                                    value={versionSelected?.schema_content}
                                />
                            )}
                            {isDiff === 'Yes' && (
                                <DiffEditor
                                    height="calc(100% - 100px)"
                                    language={schemaDetails?.type === 'protobuf' ? 'proto' : schemaDetails?.type === 'avro' ? 'json' : schemaDetails?.type}
                                    original={currentVersion?.schema_content}
                                    modified={versionSelected?.schema_content}
                                    options={{
                                        renderSideBySide: false,
                                        readOnly: true,
                                        scrollbar: { verticalScrollbarSize: 3, horizontalScrollbarSize: 0 },
                                        renderOverviewRuler: false,
                                        colorDecorators: true,
                                        fontSize: '14px',
                                        fontFamily: 'Inter'
                                    }}
                                />
                            )}
                        </>
                    )}
                    {(validateError || validateSuccess) && (
                        <div className={validateSuccess ? 'validate-note success' : 'validate-note error'}>
                            {validateError && <ErrorOutlineRounded />}
                            {validateSuccess && <CheckCircleOutlineRounded />}
                            <p>{validateError || validateSuccess}</p>
                        </div>
                    )}
                </div>
                <div className="used-stations">
                    <div className="header">
                        <p>{schemaDetails?.used_stations?.length > 0 ? 'Enforced stations' : 'Not in use'}</p>
                        <Button
                            width="130px"
                            height="30px"
                            placeholder={
                                <div className="attach-button">
                                    <AddRounded className="add" />
                                    <span>Enforce</span>
                                    {isCloud() && !state?.allowedActions?.can_enforce_schema && <LockFeature />}
                                </div>
                            }
                            radiusType="semi-round"
                            backgroundColorType="white"
                            border="gray-light"
                            onClick={() => (!isCloud() || state?.allowedActions?.can_enforce_schema) && setAttachStaionModal(true)}
                        />
                    </div>
                    {schemaDetails?.used_stations?.length > 0 && (
                        <div className="stations-list">
                            {schemaDetails.used_stations?.map((station, index) => {
                                return (
                                    <div className="station-wrapper" key={index} onClick={() => goToStation(station)}>
                                        <OverflowTip className="ovel-station" text={station} maxWidth="130px" cursor="pointer">
                                            {station}
                                        </OverflowTip>
                                        <RedirectIcon />
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>
            </div>
            <div className="footer">
                <div className="left-side">
                    <Button
                        width="105px"
                        height="34px"
                        placeholder={'Close'}
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        border="gray-light"
                        fontSize="12px"
                        fontWeight="600"
                        onClick={() => closeDrawer()}
                    />
                    {!versionSelected?.active ? (
                        <Button
                            width="115px"
                            height="34px"
                            placeholder={
                                <div className="placeholder-button">
                                    <ScrollBackIcon alt="scrollBackIcon" />
                                    <p>Activate</p>
                                </div>
                            }
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            onClick={() => setRollBackModal(true)}
                        />
                    ) : (
                        <Button
                            width="115px"
                            height="34px"
                            placeholder={'Create version'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            loading={loading}
                            disabled={!updated || (updated && newVersion === '') || validateError !== ''}
                            onClick={() => createNewVersion()}
                        />
                    )}
                </div>
            </div>
            <Modal
                header={<RollBackIcon alt="rollBackIcon" />}
                width="400px"
                height="160px"
                displayButtons={false}
                clickOutside={() => setRollBackModal(false)}
                open={rollBackModal}
            >
                <div className="roll-back-modal">
                    <p className="title">Are you sure you want to activate this version?</p>
                    <p className="desc">Your current schema will be changed to this version.</p>
                    <div className="buttons">
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Close"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => setRollBackModal(false)}
                        />
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Confirm"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            loading={rollLoading}
                            onClick={() => rollBackVersion()}
                        />
                    </div>
                </div>
            </Modal>
            <Modal
                header={<RollBackIcon alt="rollBackIcon" />}
                width="430px"
                height="200px"
                displayButtons={false}
                clickOutside={() => setActivateVersionModal(false)}
                open={activateVersionModal}
            >
                <div className="roll-back-modal">
                    <p className="title">You created a new version of the schema. Do you want to activate it?</p>
                    <p className="desc">Your schema will be updated to the chosen version.</p>
                    <div className="buttons">
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="No"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => setActivateVersionModal(false)}
                        />
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Yes"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            loading={rollLoading}
                            onClick={() => rollBackVersion(true)}
                        />
                    </div>
                </div>
            </Modal>
            <Modal
                className="attach-station-modal"
                header={
                    <div className="img-wrapper">
                        <StationsActiveIcon alt="stationsIconActive" />
                    </div>
                }
                width="400px"
                height="560px"
                hr={false}
                displayButtons={false}
                clickOutside={() => {
                    setIsUpdate(false);
                    setAttachStaionModal(false);
                }}
                open={attachStaionModal}
            >
                <AttachStationModal
                    close={() => {
                        setIsUpdate(false);
                        setAttachStaionModal(false);
                    }}
                    schemaName={schemaDetails.schema_name}
                    handleAttachedStations={updateStations}
                    attachedStations={schemaDetails?.used_stations}
                    update={isUpdate}
                />
            </Modal>
        </schema-details>
    );
}

export default SchemaDetails;
