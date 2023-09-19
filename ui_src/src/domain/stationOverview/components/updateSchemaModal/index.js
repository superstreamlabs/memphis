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

import React, { useEffect, useState } from 'react';

import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { ReactComponent as TypeIcon } from '../../../../assets/images/typeIcon.svg';
import { ReactComponent as CreatedByIcon } from '../../../../assets/images/createdByIcon.svg';
import { ReactComponent as SchemaItemIcon } from '../../../../assets/images/schemaItemIcon.svg';
import { httpRequest } from '../../../../services/http';
import Button from '../../../../components/button';
import Copy from '../../../../components/copy';
import SelectComponent from '../../../../components/select';
import Editor, { DiffEditor, loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import SegmentButton from '../../../../components/segmentButton';

loader.init();
loader.config({ monaco });

const UpdateSchemaModal = ({ stationName, dispatch, close, schemaSelected }) => {
    const [schemaDetails, setSchemaDetails] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [useschemaLoading, setUseschemaLoading] = useState(false);
    const [activeVersion, setActiveVersion] = useState();
    const [currentVersion, setCurrentversion] = useState();
    const [isDiff, setIsDiff] = useState('Yes');

    const getUpdateSchema = async () => {
        try {
            setIsLoading(true);
            const data = await httpRequest('GET', `${ApiEndpoints.GET_UPDATE_SCHEMA}?station_name=${stationName}`);
            if (data) {
                let active = data.versions?.findIndex((version) => version?.active === true);
                let current = data.versions?.findIndex((version) => version?.active === false);
                setActiveVersion(data.versions[active]);
                setCurrentversion(data.versions[current]);
                setSchemaDetails(data);
            }
        } catch (error) {}
        setIsLoading(false);
    };

    useEffect(() => {
        getUpdateSchema();
    }, []);

    const useSchema = async () => {
        try {
            setUseschemaLoading(true);
            const data = await httpRequest('POST', ApiEndpoints.USE_SCHEMA, { station_names: [stationName], schema_name: schemaSelected });
            if (data) {
                dispatch(data);
                setUseschemaLoading(false);
            }
        } catch (error) {}
        setUseschemaLoading(false);
    };

    return (
        <div className="update-schema-modal-container">
            <div className="scrollable-wrapper">
                <div className="schema-name">
                    <SchemaItemIcon />
                    <div className="name-wrapper">
                        <p className="title">Schema name</p>
                        <p className="name">{schemaSelected}</p>
                    </div>
                </div>
                <div className="type-created">
                    <div className="wrapper">
                        <TypeIcon alt="typeIcon" />
                        <p>Type:</p>
                        {schemaDetails.type === 'json' ? <p className="schema-json-name">JSON schema</p> : <span> {schemaDetails.type}</span>}
                    </div>
                    <div className="wrapper">
                        <CreatedByIcon alt="createdByIcon" />
                        <p>Created by:</p>
                        <span>{activeVersion?.created_by_username}</span>
                    </div>
                </div>
                <div className="schema-content">
                    <div className="header">
                        <div className="diff-wrapper">
                            <span>Diff : </span>
                            <SegmentButton options={['Yes', 'No']} onChange={(e) => setIsDiff(e)} />
                        </div>
                        <div className="structure-message">
                            {schemaDetails.type === 'protobuf' && (
                                <>
                                    <p className="field-name">Master message :</p>
                                    <SelectComponent
                                        value={activeVersion?.message_struct_name}
                                        colorType="black"
                                        backgroundColorType="white"
                                        borderColorType="gray-light"
                                        radiusType="semi-round"
                                        minWidth="12vw"
                                        width="100px"
                                        height="26px"
                                        options={[]}
                                        iconColor="gray"
                                        popupClassName="message-option"
                                        disabled={true}
                                    />
                                </>
                            )}
                        </div>
                        <div className="copy-icon">
                            <Copy data={activeVersion?.schema_content} />
                        </div>
                    </div>

                    {isDiff === 'No' && (
                        <Editor
                            options={{
                                minimap: { enabled: false },
                                scrollbar: { verticalScrollbarSize: 2 },
                                scrollBeyondLastLine: false,
                                roundedSelection: false,
                                formatOnPaste: true,
                                formatOnType: true,
                                readOnly: true,
                                fontSize: '14px',
                                fontFamily: 'Inter'
                            }}
                            language="proto"
                            height="calc(100% - 55px)"
                            value={activeVersion?.schema_content}
                        />
                    )}
                    {isDiff === 'Yes' && (
                        <DiffEditor
                            height="calc(100% - 55px)"
                            language="proto"
                            original={currentVersion?.schema_content}
                            modified={activeVersion?.schema_content}
                            options={{
                                renderSideBySide: false,
                                readOnly: true,
                                scrollbar: { verticalScrollbarSize: 2, horizontalScrollbarSize: 0 },
                                renderOverviewRuler: false,
                                colorDecorators: true,
                                fontSize: '14px',
                                fontFamily: 'Inter'
                            }}
                        />
                    )}
                </div>
                <div className="version-number">
                    <p>
                        Active version: <span>{activeVersion?.version_number}</span>
                    </p>
                </div>
            </div>
            <div className="buttons">
                <Button
                    width="150px"
                    height="35px"
                    placeholder="Close"
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    border="gray-light"
                    fontSize="13px"
                    fontFamily="InterSemiBold"
                    onClick={() => close()}
                />
                <Button
                    width="150px"
                    height="35px"
                    placeholder="Update"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="13px"
                    fontFamily="InterSemiBold"
                    isLoading={useschemaLoading}
                    onClick={useSchema}
                />
            </div>
        </div>
    );
};

export default UpdateSchemaModal;
