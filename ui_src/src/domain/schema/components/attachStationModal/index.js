// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useEffect, useState } from 'react';

import CheckboxComponent from '../../../../components/checkBox';
import Button from '../../../../components/button';
import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';

function AttachStationModal({ close, handleAttachedStations, attachedStations, schemaName }) {
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [allStations, setAllStations] = useState([]);
    const [attachLoader, setAttachLoader] = useState(false);
    const [indeterminate, setIndeterminate] = useState(false);

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(allStations.map((li) => li.name));
        setIndeterminate(false);
        if (isCheckAll) {
            setIsCheck([]);
        }
    };

    const handleCheckedClick = (id, checked) => {
        if (attachedStations.includes(id)) return;
        let checkedList = [];
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
            checkedList = isCheck.filter((item) => item !== id);
        }
        if (checked) {
            checkedList = [...isCheck, id];
            setIsCheck(checkedList);
        }
        setIsCheckAll(checkedList.length === allStations.length);
        setIndeterminate(!!checkedList.length && checkedList.length < allStations.length);
    };

    const getAllStations = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_ALL_STATIONS}`);
            setAllStations(res);
        } catch (err) {
            return;
        }
    };

    useEffect(() => {
        getAllStations();
    }, []);

    const attachToStation = async () => {
        setAttachLoader(true);
        try {
            const data = await httpRequest('POST', ApiEndpoints.USE_SCHEMA, { station_names: isCheck, schema_name: schemaName });
            if (data) {
                handleAttachedStations(isCheck);
                setAttachLoader(false);
                close();
            }
        } catch (error) {
            setAttachLoader(false);
        }
    };

    return (
        <div className="attach-station-content">
            <p className="title">Attach to Station</p>
            <p className="desc">Attaching a scheme to a station will force the producers to follow it</p>
            <div className="stations-list">
                <div className="header">
                    <CheckboxComponent indeterminate={indeterminate} checked={isCheckAll} id={'selectAll'} onChange={onCheckedAll} name={'selectAll'} />
                    <p>Station Name</p>
                </div>
                <div className="staion-wraper">
                    {allStations.length > 0 &&
                        allStations?.map((station, index) => {
                            return (
                                <div className="station-row" onClick={() => handleCheckedClick(station.name, isCheck.includes(station.name) ? false : true)}>
                                    <CheckboxComponent
                                        disabled={attachedStations.includes(station.name)}
                                        checked={isCheck.includes(station.name) || attachedStations.includes(station.name)}
                                        id={station.name}
                                        onChange={(e) => handleCheckedClick(e.target.id, e.target.checked)}
                                        name={station.name}
                                    />
                                    <p>{station.name}</p>
                                </div>
                            );
                        })}
                </div>
            </div>
            <div className="buttons">
                <Button
                    width="150px"
                    height="34px"
                    placeholder="Cancel"
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    border="gray-light"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    onClick={() => close()}
                />
                <Button
                    width="150px"
                    height="34px"
                    placeholder="Attach Selected"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    loading={attachLoader}
                    onClick={() => attachToStation()}
                />
            </div>
        </div>
    );
}

export default AttachStationModal;
