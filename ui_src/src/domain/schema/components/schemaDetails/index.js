// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import Editor, { DiffEditor } from '@monaco-editor/react';
import React, { useEffect, useState } from 'react';

import createdByIcon from '../../../../assets/images/createdByIcon.svg';
import rollBackIcon from '../../../../assets/images/rollBackIcon.svg';
import scrollBackIcon from '../../../../assets/images/scrollBackIcon.svg';
import SelectVersion from '../../../../components/selectVersion';
import typeIcon from '../../../../assets/images/typeIcon.svg';
import RadioButton from '../../../../components/radioButton';
import TagsList from '../../../../components/tagsList';
import Button from '../../../../components/button';
import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import Modal from '../../../../components/modal';
import { message } from 'antd';

const formatOption = [
    {
        id: 1,
        value: 0,
        label: 'Code'
    },
    {
        id: 2,
        value: 1,
        label: 'Table'
    }
];

function SchemaDetails({ schemaName, closeDrawer }) {
    const [versionSelected, setVersionSelected] = useState();
    const [currentVersion, setCurrentversion] = useState();
    const [updated, setUpdated] = useState(false);
    const [loading, setIsLoading] = useState(false);
    const [rollLoading, setIsRollLoading] = useState(false);
    const [newVersion, setNewVersion] = useState('');
    const [schemaDetails, setSchemaDetails] = useState({
        schema_name: '',
        type: '',
        version: []
    });
    const [rollBackModal, setRollBackModal] = useState(false);

    const arrangeData = (schema) => {
        let index = schema.versions?.findIndex((version) => version?.active === true);
        setCurrentversion(schema.versions[index]);
        setVersionSelected(schema.versions[index]);
        setNewVersion(schema.versions[index].schema_content);
        setSchemaDetails(schema);
    };

    const getScemaDetails = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_SCHEMA_DETAILS}?schema_name=${schemaName}`);
            arrangeData(data);
        } catch (err) {}
    };

    useEffect(() => {
        getScemaDetails();
    }, []);

    const handleSelectVersion = (e) => {
        let index = schemaDetails?.versions?.findIndex((version) => version.version_number === e);
        setVersionSelected(schemaDetails?.versions[index]);
    };

    const createNewVersion = async () => {
        try {
            setIsLoading(true);
            const data = await httpRequest('POST', ApiEndpoints.CREATE_NEW_VERSION, { schema_name: schemaName, schema_content: newVersion });
            if (data) {
                arrangeData(data);
                message.success({
                    key: 'memphisSuccessMessage',
                    content: 'New version was created successfully',
                    duration: 5,
                    style: { cursor: 'pointer' },
                    onClick: () => message.destroy('memphisSuccessMessage')
                });
                setIsLoading(false);
            }
        } catch (err) {}
        setIsLoading(false);
    };
    const rollBackVersion = async () => {
        try {
            setIsLoading(true);
            const data = await httpRequest('PUT', ApiEndpoints.ROLL_BACK_VERSION, { schema_name: schemaName, version_number: versionSelected?.version_number });
            if (data) {
                arrangeData(data);
                setRollBackModal(false);
            }
        } catch (err) {}
        setIsLoading(false);
    };

    return (
        <schema-details is="3xd">
            <div className="scrollable-wrapper">
                <div className="type-created">
                    <div className="wrapper">
                        <img src={typeIcon} />
                        <p>Type:</p>
                        <span>{schemaDetails?.type}</span>
                    </div>
                    <div className="wrapper">
                        <img src={createdByIcon} />
                        <p>Created by:</p>
                        <span>{currentVersion?.created_by_user}</span>
                    </div>
                </div>
                <div className="tags">
                    <TagsList tags={schemaDetails?.tags} addNew={true} />
                </div>
                <div className="schema-fields">
                    <div className="left">
                        <p>Schema</p>
                        {/* <RadioButton options={formatOption} radioValue={passwordType} onChange={(e) => passwordTypeChange(e)} /> */}
                    </div>
                    <SelectVersion value={versionSelected?.version_number} options={schemaDetails?.versions} onChange={(e) => handleSelectVersion(e)} />
                </div>
                <div className="schema-content">
                    {versionSelected?.active && (
                        <Editor
                            options={{
                                minimap: { enabled: false },
                                scrollbar: { verticalScrollbarSize: 0 },
                                scrollBeyondLastLine: false,
                                roundedSelection: false,
                                formatOnPaste: true,
                                formatOnType: true
                            }}
                            language="proto"
                            defaultValue={versionSelected?.schema_content}
                            value={newVersion}
                            onChange={(value) => {
                                setUpdated(true);
                                setNewVersion(value);
                            }}
                        />
                    )}
                    {!versionSelected?.active && (
                        <DiffEditor
                            height="90%"
                            language="proto"
                            original={currentVersion?.schema_content}
                            modified={versionSelected?.schema_content}
                            options={{
                                renderSideBySide: false,
                                scrollbar: { verticalScrollbarSize: 0, horizontalScrollbarSize: 0 },
                                scrollBeyondLastLine: false,
                                scrollBeyondLastColumn: false
                            }}
                        />
                    )}
                    {updated && newVersion === '' && (
                        <div>
                            <p className="warning-message">Schema content cannot be empty</p>
                        </div>
                    )}
                </div>
                <div className="used-stations">
                    {schemaDetails?.used_stations?.length > 0 ? (
                        <>
                            <p className="title">Used by stations</p>
                            <div className="stations-list">
                                {schemaDetails.used_stations?.map((station, index) => {
                                    return (
                                        <div className="station-wrapper" key={index}>
                                            <p>{station}</p>
                                        </div>
                                    );
                                })}
                            </div>
                        </>
                    ) : (
                        <p className="title">Not Used yet</p>
                    )}
                </div>
            </div>
            <div className="footer">
                <div>
                    {!versionSelected?.active && (
                        <Button
                            width="111px"
                            height="34px"
                            placeholder={
                                <div className="placeholder-button">
                                    <img src={scrollBackIcon} />
                                    <p>Roll back</p>
                                </div>
                            }
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            onClick={() => setRollBackModal(true)}
                        />
                    )}
                </div>
                <div className="left-side">
                    <Button
                        width="111px"
                        height="34px"
                        placeholder={'close'}
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        border="gray-light"
                        fontSize="12px"
                        fontWeight="600"
                        onClick={() => closeDrawer()}
                    />
                    <Button
                        width="111px"
                        height="34px"
                        placeholder={'Update'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        loading={loading}
                        disabled={!updated || (updated && newVersion === '')}
                        onClick={() => createNewVersion()}
                    />
                </div>
            </div>
            <Modal
                header={<img src={rollBackIcon} />}
                width="400px"
                height="160px"
                displayButtons={false}
                clickOutside={() => setRollBackModal(false)}
                open={rollBackModal}
            >
                <div className="roll-back-modal">
                    <p className="title">Are you sure you want to roll back?</p>
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
        </schema-details>
    );
}

export default SchemaDetails;
