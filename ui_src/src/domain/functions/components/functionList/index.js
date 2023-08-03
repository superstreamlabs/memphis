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

import React, { useEffect, useContext, useState } from 'react';
import placeholderFunctions from '../../../../assets/images/placeholderFunctions.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Loader from '../../../../components/loader';
import Button from '../../../../components/button';
import Filter from '../../../../components/filter';
import { Context } from '../../../../hooks/store';
import { useHistory } from 'react-router-dom';
import FunctionBox from '../functionBox';

function FunctionList({ createNew }) {
    const history = useHistory();
    const [state, dispatch] = useContext(Context);
    const [isLoading, setisLoading] = useState(true);

    useEffect(() => {
        getAllSchemas();
        return () => {
            dispatch({ type: 'SET_SCHEMA_LIST', payload: [] });
            dispatch({ type: 'SET_STATION_FILTERED_LIST', payload: [] });
        };
    }, []);

    const getAllSchemas = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            dispatch({ type: 'SET_SCHEMA_LIST', payload: data });
            dispatch({ type: 'SET_STATION_FILTERED_LIST', payload: data });
            setTimeout(() => {
                setisLoading(false);
            }, 500);
        } catch (error) {
            setisLoading(false);
        }
    };

    return (
        <div className="schema-container">
            <div className="header-wraper">
                <div className="main-header-wrapper">
                    <label className="main-header-h1">
                        Functions <label className="length-list">{state.schemaFilteredList?.length > 0 && `(${state.schemaFilteredList?.length})`}</label>
                    </label>
                    <span className="memphis-label">Serverless functions to process ingested events "on the fly"</span>
                </div>
                <div className="action-section">
                    <Filter filterComponent="schemaverse" height="34px" />
                    <Button
                        width="160px"
                        height="34px"
                        placeholder={'Create from blank'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        boxShadowStyle="float"
                        aria-haspopup="true"
                        onClick={() => {}}
                    />
                </div>
            </div>

            <div className="schema-list">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {!isLoading &&
                    state.schemaFilteredList?.map((func, index) => {
                        return <FunctionBox key={index} funcDetails={func} />;
                    })}
                {!isLoading && state.schemaList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <img src={placeholderFunctions} width="150" alt="placeholderFunctions" />
                        <p className="title">No functions yet</p>
                        <p className="sub-title">Functions will start to sync and appear once an integration with a git repository is established.</p>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="34px"
                            placeholder="Integrate to Gitlab"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            aria-controls="usecse-menu"
                            aria-haspopup="true"
                            onClick={() => {}}
                        />
                    </div>
                )}
                {!isLoading && state.schemaList?.length > 0 && state.schemaFilteredList?.length === 0 && (
                    <div className="no-schema-to-display">
                        <img src={placeholderFunctions} width="150" alt="placeholderFunctions" />
                        <p className="title">No functions found</p>
                        <p className="sub-title">Please try to search again</p>
                    </div>
                )}
            </div>
        </div>
    );
}

export default FunctionList;
