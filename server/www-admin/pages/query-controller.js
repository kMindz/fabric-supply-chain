/**
 * Created by maksim on 7/13/17.
 */
/**
 * @class QueryController
 * @ngInject
 */
function QueryController($scope, ChannelService, ConfigLoader, $log, $q) {
  var ctl = this;

  ctl.channels = [];
  ctl.chaincodes = [];
  ctl.transaction = null;
  ctl.invokeInProgress = false;
  ctl.EditProcess = false;
  ctl.arg = [];
  ctl.MyProducts = [];
  ctl.AllProducts = [];
  ctl.HistoryProduct = [];
  ctl.showHistory=false;
  ctl.Orgs = ConfigLoader.getOrgs();

  ctl.Obj =[];

  // init
  var orgs = ConfigLoader.getOrgs();

  ctl.States = ConfigLoader.getStates();
  ctl.channel  = Object.create({}, { channel_id: { value: 'common'} });
  ctl.chaincode =  Object.create({}, { name: { value: 'reference'} });
  ctl.arg['Org']   = ConfigLoader.get().org;
  ctl.peers = [ctl.arg['Org']+'/peer0', ctl.arg['Org']+'/peer1'];

  $scope.sortType     = 'name'; // set the default sort type
  $scope.sortReverse  = false;  // set the default sort order
  $scope.searchFish   = '';

  var allPeers = [];
  orgs.forEach(function(org){
    var peers = ConfigLoader.getPeers(org.id);
    allPeers.push.apply(allPeers, peers);
  });


  ctl.getPeers = function(){
    return allPeers;
  };

  ctl.getChannels = function(){
    return ChannelService.list().then(function(dataList){
      ctl.channels = dataList;
    });
  };

  ctl.getChaincodes = function(){
    if(!$scope.selectedChannel){
      return $q.resolve([]);
    }
    return ChannelService.listChannelChaincodes($scope.selectedChannel.channel_id).then(function(dataList){
      ctl.chaincodes = dataList;
    });
  };

  ctl.add = function(product, description){

    ctl.arg['Name']  = product;
    ctl.arg['Desc']  = description;
    ctl.arg['State'] = 1;
    ctl.arg['Time']  = Math.floor(Date.now() / 1000);

    ctl.fcn ='initProduct';

    ctl.invoke(ctl.channel, ctl.chaincode, ctl.peers, ctl.fcn, '["'+ctl.arg['Name']+'","'+ctl.arg['Desc']+'", "'+ctl.arg['State']+'", "'+ctl.arg['Org']+'", "'+ctl.arg['Time']+'"]');
  };

  ctl.edit = function(name, description, state, owner){

    ctl.arg['Name']  = name;
    ctl.arg['Desc']  = description;
    ctl.arg['State'] = getIdState(state.name);
    ctl.arg['Owner'] = owner;
    ctl.arg['Time']  = Math.floor(Date.now() / 1000);
    ctl.fcn ='updateProduct';

    ctl.invoke(ctl.channel, ctl.chaincode, ctl.peers, ctl.fcn, '["'+ctl.arg['Name']+'","'+ctl.arg['Desc']+'", "'+ctl.arg['State']+'","'+ctl.arg['Owner']+'", "'+ctl.arg['Time']+'"]');

  }

  ctl.delete = function(name){

     ctl.arg['Name']  = name;
     ctl.fcn ='delete';
     ctl.invoke(ctl.channel, ctl.chaincode, ctl.peers, ctl.fcn, '["'+ctl.arg['Name']+'"]');

  }

  ctl.my = function(args){
     ctl.fcn ='queryProductsByOwner';
     ctl.query(ctl.channel, ctl.chaincode, ctl.arg['Org']+'/peer0', ctl.fcn, '["'+args+'"]');
  }

  ctl.all = function(args){
     ctl.fcn ="queryProducts";
     var docType = "product";
     var owner1   = 'a';
     var owner2   = 'b';
     var owner3   = 'c';
     var result=JSON.stringify({
          "selector": {
              "docType": docType,
              "owner"  : {
                "$in": [owner1, owner2, owner3]
              } ,
          },
     });
     ctl.query(ctl.channel, ctl.chaincode, ctl.arg['Org']+'/peer0', ctl.fcn, '["'+encodeURI(result)+'"]');
  }

    ctl.history = function(name){
        ctl.arg['Name']  = name;
        ctl.fcn ='getHistoryForProduct';
        ctl.invoke(ctl.channel, ctl.chaincode, ctl.peers, ctl.fcn, '["'+ctl.arg['Name']+'"]');
    }

  ctl.invoke = function(channel, cc, peers, fcn, args){
    try{
      args = JSON.parse(args);
    }catch(e){
      $log.warn(e);
    }

    ctl.transaction = null;
    ctl.error = null;
    ctl.invokeInProgress = true;

    return ChannelService.invoke(channel.channel_id, cc.name, peers, fcn, args)
      .then(function(data){
        return ChannelService.getTransactionById(channel.channel_id, data.transaction);
      })
      .then(function(transaction){
        ctl.transaction = transaction;
        ctl.result = getTxResult(transaction);
          if(ctl.fcn ==="getHistoryForProduct"){
              ctl.HistoryProduct = JSON.parse(transaction.transactionEnvelope.payload.data.actions[0].payload.action.proposal_response_payload.extension.response.payload);
              console.info(ctl.HistoryProduct);
          }
      })
      .catch(function(response){
        ctl.error = response.data || response;
      })
      .finally(function(transaction){
        ctl.invokeInProgress = false;
        ctl.all();
        ctl.my(ctl.arg['Org']);
      });
  }

  function getTxResult(transaction){
    var result = null;
    try{
      result = {};
      // TODO: loop trough actions
      var ns_rwset = transaction.transactionEnvelope.payload.data.actions[0].payload.action.proposal_response_payload.extension.results.ns_rwset;
      ns_rwset = ns_rwset.filter(function(action){return action.namespace != "lscc"}); // filter system chaincode
      ns_rwset.forEach(function(action){
        result[action.namespace] = action.rwset.writes.reduce(function(result, element){
          result[element.key] = element.is_delete ? null : element.value;
          return result;
        }, {});

      });
    }catch(e){
      console.info(e);
      result = null
    }
    return result;
  }

  function getQTxResult(transaction){
    if(typeof transaction.result !== 'undefined' && transaction.result !== null){
        var buffer =[];
        buffer = transaction.result;
        // console.info(buffer);

        if(ctl.fcn ==="queryProductsByOwner"){
            ctl.MyProducts = [];
            buffer.forEach(function(el){
                ctl.MyProducts.push(el.Record);
            });
            ctl.MyProducts.map(function (el) {
                el.lastUpdated = (new Date(parseInt(el.lastUpdated) *1000)).toLocaleString();
                el.state = getNameState(el.state);
            });

        } else if( ctl.fcn ==="queryProducts" ){
            ctl.AllProducts = [];
            buffer.forEach(function(el){
                ctl.AllProducts.push(el.Record);
            });
            ctl.AllProducts.map(function (el) {
                el.lastUpdated = (new Date(parseInt(el.lastUpdated) *1000)).toLocaleString();
                el.state = getNameState(el.state);
            });
        }

    }

    return transaction;
  }

  function getIdState(state){
    var index;
    for(index = 0; index <ctl.States.length; ++index){
      if(ctl.States[index].name === state){
        return ctl.States[index].id;
      }
    }
    return 0;
  }

    function getNameState(state){
        var index;
        for(index = 0; index <ctl.States.length; ++index){
            if(ctl.States[index].id === state){
                return ctl.States[index].name;
            }
        }
        return 0;
    }


  ctl.query = function(channel, cc, peer, fcn, args){
    try{
      args = JSON.parse(args);
    }catch(e){
      $log.warn(e);
    }

    ctl.transaction = null;
    ctl.error = null;
    ctl.invokeInProgress = true;

    return ChannelService.query(channel.channel_id, cc.name, peer, fcn, args)
      .then(function(transaction){
        ctl.transaction = transaction;
        ctl.result = getQTxResult(transaction);
      })
      .catch(function(response){
        ctl.error = response.data || response;
      })
      .finally(function(){
        ctl.invokeInProgress = false;
      });
  }


  //
  ctl.getChannels();
  $scope.$watch('selectedChannel', ctl.getChaincodes );




}


angular.module('nsd.controller.query', ['nsd.service.channel'])
  .controller('QueryController', QueryController);
