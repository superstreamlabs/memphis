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
import attachedPlaceholder from '../../../../assets/images/attachedPlaceholder.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Button from '../../../../components/button';

function AttachStationModal({ close, handleAttachedStations, attachedStations, schemaName }) {
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [allStations, setAllStations] = useState([]);
    const [attachLoader, setAttachLoader] = useState(false);
    const [indeterminate, setIndeterminate] = useState(false);

    const onCheckedAll = (e) => {
        if (attachedStations?.length > 0) {
            setIndeterminate(!indeterminate);
            if (indeterminate) {
                setIsCheck([]);
            } else {
                allStations?.map((li) => {
                    if (attachedStations?.includes(li.name)) return;
                    else setIsCheck(...li.name);
                });
            }
        } else {
            setIsCheckAll(!isCheckAll);
            setIndeterminate(false);
            if (isCheckAll) {
                setIsCheck([]);
            } else {
                setIsCheck(allStations?.map((li) => li.name));
            }
        }
    };

    const handleCheckedClick = (id, checked) => {
        if (attachedStations?.includes(id)) return;
        let checkedList = [];
        if (!checked) {
            setIsCheck(isCheck?.filter((item) => item !== id));
            checkedList = isCheck?.filter((item) => item !== id);
        }
        if (checked) {
            checkedList = [...isCheck, id];
            setIsCheck(checkedList);
        }
        setIsCheckAll(checkedList?.length === allStations?.length);
        setIndeterminate(!!checkedList?.length && checkedList?.length < allStations?.length);
    };

    const getAllStations = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_ALL_STATIONS}`);
            let native_staion = res.filter((station) => station.is_native);
            setAllStations(native_staion);
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
                handleAttachedStations([...attachedStations, ...isCheck]);
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
                {allStations?.length > 0 ? (
                    <div className="header">
                        <CheckboxComponent
                            disabled={attachedStations?.length === allStations?.length}
                            indeterminate={indeterminate}
                            checked={isCheckAll}
                            id={'selectAll'}
                            onChange={onCheckedAll}
                            name={'selectAll'}
                        />
                        <p>Station Name</p>
                    </div>
                ) : (
                    <div className="placeholder">
                        <img src={attachedPlaceholder} alt="attachedPlaceholder" />
                        <p>No Station found</p>
                    </div>
                )}
                {allStations?.length > 0 && (
                    <div className="station-wraper">
                        {allStations?.map((station, index) => {
                            return (
                                <div
                                    key={station.name}
                                    className="station-row"
                                    onClick={() => handleCheckedClick(station.name, isCheck.includes(station.name) ? false : true)}
                                >
                                    <CheckboxComponent
                                        disabled={attachedStations?.includes(station.name)}
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
                )}
            </div>
            <div className="buttons">
                {allStations?.length > 0 && (
                    <>
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
                            disabled={isCheck?.length === 0}
                            onClick={() => attachToStation()}
                        />
                    </>
                )}
            </div>
        </div>
    );
}

export default AttachStationModal;
