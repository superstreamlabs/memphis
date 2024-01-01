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

import React, { useEffect, useState, useRef, useContext } from 'react';
import { AddRounded } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import pathDomains from 'router';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import { ReactComponent as SearchIcon } from 'assets/images/searchIcon.svg';
import SearchInput from 'components/searchInput';
import Button from 'components/button';
import Modal from 'components/modal';
import Loader from 'components/loader';
import LearnMore from 'components/learnMore';
import SchemaItem from '../../../stationOverview/components/useSchemaModal/schemaItem';
import { ReactComponent as StationIcon } from 'assets/images/stationsIconActive.svg';
import CreateStationForm from 'components/createStationForm';
import { isCloud } from 'services/valueConvertor';
import { Context } from 'hooks/store';
import LockFeature from 'components/lockFeature';

const AttachFunctionModal = ({ open, clickOutside, selectedFunction }) => {
    const createStationRef = useRef(null);
    const [state, dispatch] = useContext(Context);
    const [searchInput, setSearchInput] = useState('');
    const [isLoading, setIsLoading] = useState(true);
    const [selected, setSelected] = useState(null);
    const [newStationModal, setNewStationModal] = useState(false);
    const [creatingProsessd, setCreatingProsessd] = useState(false);
    const [allStations, setAllStations] = useState([]);
    const [filteredStations, setFilteredStations] = useState([]);
    const history = useHistory();

    useEffect(() => {
        open && getAllStations();
    }, [open]);

    useEffect(() => {
        if (searchInput === '') {
            setFilteredStations(allStations);
        } else {
            setFilteredStations(allStations?.filter((station) => station.name.toLowerCase().includes(searchInput.toLowerCase())));
        }
    }, [allStations, searchInput]);

    const getAllStations = async () => {
        setIsLoading(true);
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_ALL_STATIONS}`);
            let native_station = res.filter((station) => station.is_native)?.filter((station) => station?.version >= 2);
            setAllStations(native_station);
            setIsLoading(false);
        } catch (err) {
            setIsLoading(false);
        }
    };

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    return (
        <Modal
            header={
                <div className="modal-header">
                    <p>Attach a function</p>
                </div>
            }
            displayButtons={false}
            height="400px"
            width="352px"
            clickOutside={clickOutside}
            open={open}
            hr={true}
            className="use-schema-modal"
        >
            <div className="attach-station-modal-container">
                {isLoading && <Loader />}
                {!isLoading && filteredStations?.length === 0 && (
                    <div className="no-schema-to-display">
                        <StationIcon width={50} height={50} />

                        <p className="title">No stations yet</p>
                        <p className="sub-title">Get started by creating your first station</p>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="34px"
                            placeholder={`Create new station`}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            aria-controls="usecse-menu"
                            aria-haspopup="true"
                            onClick={() => setNewStationModal(true)}
                        />
                    </div>
                )}
                {!isLoading && filteredStations?.length > 0 && (
                    <>
                        <SearchInput
                            placeholder={`Search station`}
                            colorType="navy"
                            backgroundColorType="none"
                            borderRadiusType="circle"
                            borderColorType="search-input"
                            iconComponent={<SearchIcon />}
                            onChange={handleSearch}
                            value={searchInput}
                            width="100%"
                            height="35px"
                        />
                        <div className="schemas-list">
                            {filteredStations?.map((station) => {
                                return <SchemaItem key={station.name} schema={station} selected={selected} handleSelectedItem={(id) => setSelected(id)} type={'dls'} />;
                            })}
                        </div>

                        <div className="buttons">
                            <div className="add-schema" onClick={() => (!isCloud() || state?.allowedActions?.can_create_stations) && setNewStationModal(true)}>
                                <AddRounded />
                                <p>Add a new station </p>
                                {isCloud() && !state?.allowedActions?.can_create_stations && <LockFeature />}
                            </div>
                            <Button
                                width="100%"
                                height="35px"
                                placeholder="Attach"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="13px"
                                fontFamily="InterSemiBold"
                                disabled={!selected}
                                onClick={() =>
                                    history.push({
                                        pathname: `${pathDomains.stations}/${selected}`,
                                        selectedFunction: selectedFunction
                                    })
                                }
                            />
                        </div>
                    </>
                )}

                <Modal
                    header={
                        <div className="modal-header">
                            <div className="header-img-container">
                                <StationIcon className="headerImage" alt="stationImg" />
                            </div>
                            <p>Create a new station</p>
                            <label>
                                A station is a distributed unit that stores the produced data{' '}
                                <LearnMore url="https://docs.memphis.dev/memphis/memphis-broker/concepts/station" />
                            </label>
                        </div>
                    }
                    height="58vh"
                    width="1020px"
                    rBtnText="Create"
                    lBtnText="Cancel"
                    lBtnClick={() => {
                        setNewStationModal(false);
                    }}
                    rBtnClick={() => {
                        createStationRef.current();
                        setNewStationModal(false);
                    }}
                    clickOutside={() => setNewStationModal(false)}
                    open={newStationModal}
                    isLoading={creatingProsessd}
                >
                    <CreateStationForm
                        createStationFormRef={createStationRef}
                        setLoading={(e) => setCreatingProsessd(e)}
                        finishUpdate={() => {
                            getAllStations();
                        }}
                        noRedirect={true}
                    />
                </Modal>
            </div>
        </Modal>
    );
};

export default AttachFunctionModal;
